## Basic worker with graceful shutdown

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("Hello, playground")
	mgr := NewTaskManager()
	// Start 3 worker.
	for i := 0; i < 3; i++ {
		mgr.Start()
	}

	go func() {
		// Send message to workers.
		for i := 100; i < 200; i++ {
			msg := fmt.Sprintf("work %d", i)
			if sent := mgr.Send(msg); sent {
				fmt.Println("sent", msg)
			}
		}
	}()

	go func() {
		// Send message to workers.
		for i := 0; i < 100; i++ {
			msg := fmt.Sprintf("work %d", i)
			if sent := mgr.Send(msg); sent {
				fmt.Println("sent", msg)
			}
		}
	}()
	time.Sleep(2 * time.Second)
	mgr.Stop()
	fmt.Println("terminating")
}

type TaskManager struct {
	ch   chan interface{}
	done sync.WaitGroup
	quit chan interface{}
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		ch:   make(chan interface{}),
		quit: make(chan interface{}),
	}
}

func (t *TaskManager) Start() {
	t.done.Add(1)
	go func() {
		defer t.done.Done()
		for {
			select {
			case <-t.quit:
				fmt.Println("quitting", len(t.ch))
				return
			case evt, ok := <-t.ch:
				if !ok {
					return
				}
				time.Sleep(1 * time.Second)
				fmt.Println("received evt", evt)
			}
		}
	}()
}

func (t *TaskManager) Stop() {
	fmt.Println("stopping")

	close(t.quit)
	t.done.Wait()

	fmt.Println("stopped")
}

func (t *TaskManager) Send(evt interface{}) bool {
	// Add a consumer task to keep track of messages sent.

	select {
	case <-t.quit:
		return false
	case t.ch <- evt:
		return true
	case <-time.After(5 * time.Second):
		return false
	}
}

```

## Multiple Background Worker implementation

```go
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Response interface {
	Value() interface{}
	Error() error
}

type Task interface {
	Execute() Response
}

type Delay struct {
	Duration int
	Err      error
}

func (d *Delay) Value() interface{} {
	return d.Duration
}

func (d *Delay) Error() error {
	return d.Err
}

type DelayTask struct {
	value Delay
}

func (d *DelayTask) Execute() Response {
	duration := rand.Intn(500)
	time.Sleep(time.Duration(duration) * time.Millisecond)
	return &Delay{
		Duration: duration,
	}

}

type WorkerPool struct {
	taskCh chan Task
	quit   chan interface{}
	wg     *sync.WaitGroup
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{
		// Adding a buffer allows the taskCh to take more than it can process.
		taskCh: make(chan Task, 10),
		quit:   make(chan interface{}),
		wg:     new(sync.WaitGroup),
	}
}

func (w *WorkerPool) Start(n int) func() {
	w.wg.Add(n)
	for i := 0; i < n; i++ {
		go w.loop(i)
	}
	return func() {
		w.wg.Wait()
	}
}

func (w *WorkerPool) loop(i int) {
	defer w.wg.Done()
	fmt.Println("started worker", i)
	for {
		select {
		case <-w.quit:
			return
		case task, ok := <-w.taskCh:
			if !ok {
				return
			}
			res := task.Execute()
			duration, ok := res.Value().(int)
			if ok {
				fmt.Println("worker", i, "took", duration, "ms")
			}
		}
	}
}

func (w *WorkerPool) Send(task Task) {
	select {
	case <-w.quit:
		return
	case w.taskCh <- task:
	}
}

func (w *WorkerPool) Stop() {
	close(w.quit)
}

func main() {

	pool := NewWorkerPool()
	wait := pool.Start(5)
	go func() {
    // We can listen for events from a message queue for example,
    // and stream in new Tasks to the background worker.
		for i := 0; i < 100; i++ {
			pool.Send(&DelayTask{})
		}
	}()

	go func() {
		time.Sleep(10 * time.Second)
		fmt.Println("stopping")
		pool.Stop()
	}()

	wait()
}
```


## Worker with Pipeline

```go
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Result struct {
	Response interface{}
	Err      error
}

type Task interface {
	Execute() Result
}

type WorkerPool struct {
	wg     *sync.WaitGroup
	mu     *sync.RWMutex
	once   *sync.Once
	quit   chan interface{} // Own quit channel.
	busCh  chan interface{} // Global quit channel.
	taskCh chan Task

	outCh chan Result

	counter int
}

func NewWorkerPool(busCh chan interface{}, taskLimit int) *WorkerPool {
	return &WorkerPool{
		busCh:  busCh,
		mu:     new(sync.RWMutex),
		quit:   make(chan interface{}),
		taskCh: make(chan Task, taskLimit),
		once:   new(sync.Once),
		wg:     new(sync.WaitGroup),
		outCh:  make(chan Result),
	}
}

func (w *WorkerPool) Start(n int) *sync.WaitGroup {
	w.wg.Add(n)
	for i := 0; i < n; i++ {
		go w.loop()
	}
	fmt.Printf("started %d workers\n", n)
	return w.wg
}

func (w *WorkerPool) AddWorker() {
	w.mu.Lock()
	w.wg.Add(1)
	go w.loop()
	w.mu.Unlock()
	fmt.Println("add 1 worker")
}

func (w *WorkerPool) RemoveWorker() {
	w.quit <- struct{}{}
}

func (w *WorkerPool) AddTask(tasks ...Task) {
	for _, task := range tasks {
		select {
		case <-w.busCh:
			return
		case <-w.quit:
			return
		case w.taskCh <- task:
		}
	}
}

func (w *WorkerPool) loop() {
	defer func() {
		fmt.Println("worker stopped")
		w.wg.Done()
	}()
	for {
		select {
		case <-w.busCh:
			return
		case <-w.quit:
			return
		case task, ok := <-w.taskCh:
			if !ok {
				return
			}
			res := task.Execute()

			w.mu.Lock()
			w.counter++
			w.mu.Unlock()

			fmt.Println("task:", res.Response)
			select {
			case <-w.quit:
				return
			case w.outCh <- res:
			}
		}
	}

}
func (w *WorkerPool) Stop() {
	w.once.Do(func() {
		close(w.quit)
	})
}

func main() {
	done := make(chan interface{})
	pool := NewWorkerPool(done, 100)

	numWorkers := 1

	// Create a new worker pool with n workers.
	job := pool.Start(numWorkers)

	go generator(pool, 100)

	go func() {
		time.Sleep(1 * time.Second)

		for i := 0; i < 5; i++ {
			pool.AddWorker()
		}
		pool.RemoveWorker()
		pool.RemoveWorker()
	}()
	
	reader := func() {
		// We can chain the pipeline.
		for res := range multiplier(done, pool.outCh, 2) {
			fmt.Println("output", res)
		}
	}
	go reader()
	go func() {
		time.Sleep(5 * time.Second)
		close(done)
	}()
	job.Wait()
	fmt.Println("exiting", pool.counter)
}

type DelayTask struct{}

func (d *DelayTask) Execute() Result {
	time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
	return Result{
		Response: 2,
	}
}

// Pipelines
func multiplier(
	done chan interface{},
	inCh <-chan Result,
	m int,
) <-chan interface{} {
	outCh := make(chan interface{})
	go func() {
		defer close(outCh)
		for v := range inCh {
			i, ok := v.Response.(int)
			if !ok {
				continue
			}
			select {
			case <-done:
				return
			case outCh <- i * 2:
			}
		}
	}()
	return outCh
}

func generator(pool *WorkerPool, n int) {
	for i := 0; i < n; i++ {
		go func() {
			time.Sleep(time.Duration(rand.Intn(500)+250) * time.Millisecond)
			pool.AddTask(&DelayTask{})
		}()
	}
}
```
