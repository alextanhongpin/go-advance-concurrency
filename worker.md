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
