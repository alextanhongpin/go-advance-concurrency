# Basic
```golang
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	c := sync.NewCond(&sync.Mutex{})

	var ok bool
	conditionTrue := func() bool {
		return ok
	}

	go func() {
		time.Sleep(1 * time.Second)
		ok = true
		fmt.Println("broadcasting")
		c.Broadcast()
	}()

	var wg sync.WaitGroup
	wg.Add(2)
	doWork := func(i int) {
		defer wg.Done()
		c.L.Lock()
		for conditionTrue() == false {
			fmt.Println("waiting", i)
			c.Wait()
			fmt.Println("done", i)
		}
		c.L.Unlock()
	}

	go doWork(1)
	go doWork(2)
	wg.Wait()

	fmt.Println("exiting")
}
```

Output:

```
waiting 2
waiting 1
broadcasting
done 1
done 2
exiting
```

## Ensuring the queue cannot be more than 2

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	c := sync.NewCond(&sync.Mutex{})
	queue := make([]interface{}, 0, 10)

	removeFromQueue := func(delay time.Duration) {
		time.Sleep(delay)
		c.L.Lock()
		queue = queue[1:]
		fmt.Println("removed from queue")
		c.L.Unlock()
		c.Signal()
	}

	for i := 0; i < 10; i++ {
		c.L.Lock()
		for len(queue) == 2 {
			// Wait does not block - it suspends the main 
			// goroutine until a signal on the condition has been sent.
			c.Wait()
		}
		fmt.Println("adding to queue:", i)
		queue = append(queue, struct{}{})
		go removeFromQueue(1 * time.Second)
		c.L.Unlock()
	}
}
```

The example above can be rewritten with channels of course:

```golang
package main

import (
	"fmt"
	"time"
)

func main() {
	done := make(chan struct{})
	defer close(done)

	generator := func(done chan struct{}, nums ...int) <-chan interface{} {
		outStream := make(chan interface{})
		go func() {
			defer close(outStream)
			for _, v := range nums {
				outStream <- v
			}
		}()
		return outStream
	}
	queue := func(done chan struct{}, in <- chan interface{}, limit int)<-chan interface{} {
		fmt.Println("buffer set to", limit)
		outStream := make(chan interface{}, limit)
		go func() {
			defer close(outStream)
			for v := range in {
				fmt.Println("buffer", v, len(outStream))
				outStream <- v
			}
		}()
		return outStream
	}

	doWork := func(done chan struct{}, in <-chan interface{}) <-chan interface{} {
		outStream := make(chan interface{})
		go func() {
			defer close(outStream)
			for v := range in {
				fmt.Println("read", v)
				time.Sleep(1 * time.Second)
				outStream <- v
			}

		}()
		return outStream
	}
	
	genNums := generator(done, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	result := doWork(done, queue(done, genNums, 2))
	for v := range result {
		fmt.Println(v)
	}
}
```


## Worker Pool implementation


Initially I thought that using `sync.Cond` would be helpful, but it is not necessary. See the implementation below:
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
	mu     *sync.Mutex
	cond   *sync.Cond
	once   *sync.Once
	quit   chan interface{}
	taskCh chan Task

	counter int
}

func NewWorkerPool(taskLimit int) *WorkerPool {
	return &WorkerPool{
		mu:     new(sync.Mutex),
		quit:   make(chan interface{}),
		taskCh: make(chan Task, taskLimit),
		cond:   sync.NewCond(new(sync.Mutex)),
		once:   new(sync.Once),
		wg:     new(sync.WaitGroup),
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

func (w *WorkerPool) AddTask(tasks ...Task) {
	for _, task := range tasks {
		select {
		case <-w.quit:
			return
		case w.taskCh <- task:
			w.cond.Broadcast()
			fmt.Println("received task", task)
		}
	}
}

func (w *WorkerPool) loop() {
	defer w.wg.Done()
	for {
		select {
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
			fmt.Println("task:", res)
		default:
			w.cond.L.Lock()
			w.cond.Wait()
			w.cond.L.Unlock()
		}
	}

}
func (w *WorkerPool) Stop() {
	w.once.Do(func() {
		close(w.quit)
		w.cond.Broadcast()
	})
}

func main() {
	wp := NewWorkerPool(100)

	numWorkers := 5
	job := wp.Start(numWorkers)
	go func() {
		for i := 0; i < 100; i++ {
			wp.AddTask(&DelayTask{})
		}
	}()
	go func() {
		time.Sleep(5 * time.Second)
		wp.Stop()
	}()

	job.Wait()
	fmt.Println("exiting", wp.counter)
}

type DelayTask struct{}

func (d *DelayTask) Execute() Result {
	time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
	return Result{
		Response: "done",
	}
}
```


without `sync.Cond`:

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
	quit   chan interface{}
	taskCh chan Task

	counter int
}

func NewWorkerPool(taskLimit int) *WorkerPool {
	return &WorkerPool{
		mu:     new(sync.RWMutex),
		quit:   make(chan interface{}),
		taskCh: make(chan Task, taskLimit),
		once:   new(sync.Once),
		wg:     new(sync.WaitGroup),
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

func (w *WorkerPool) AddTask(tasks ...Task) {
	for _, task := range tasks {
		select {
		case <-w.quit:
			return
		case w.taskCh <- task:
		}
	}
}

func (w *WorkerPool) loop() {
	defer w.wg.Done()
	for {
		select {
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
		}
		fmt.Println("looping")
	}

}
func (w *WorkerPool) Stop() {
	w.once.Do(func() {
		close(w.quit)
	})
}

func main() {
	wp := NewWorkerPool(100)

	numWorkers := 10
	job := wp.Start(numWorkers)
	go func() {
		for i := 0; i < 100; i++ {
			go func() {
				time.Sleep(time.Duration(rand.Intn(500)+250) * time.Millisecond)
				wp.AddTask(&DelayTask{})
			}()

		}
	}()
	go func() {
		time.Sleep(5 * time.Second)
		wp.Stop()
	}()

	job.Wait()
	fmt.Println("exiting", wp.counter)
}

type DelayTask struct{}

func (d *DelayTask) Execute() Result {
	time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
	return Result{
		Response: "done",
	}
}
```

## Conditional counting


In the implementation below, the goroutines will continue locking until the conditions are fulfilled.


```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {

	cond := sync.NewCond(&sync.Mutex{})

	var wg sync.WaitGroup
	wg.Add(2)

	var n int

	go func() {
		defer wg.Done()

		cond.L.Lock()

		for n != 5 {
			fmt.Println("thread #1", n)
			cond.Wait()
		}
		fmt.Println("thread #1: done")
		cond.L.Unlock()
	}()

	go func() {
		defer wg.Done()

		cond.L.Lock()
		// As long as the condition is not fulfilled, the thread remains locked.
		for n != 10 {
			fmt.Println("thread #2:", n)
			cond.Wait()
		}
		fmt.Println("thread #2: done")
		cond.L.Unlock()
	}()

	go func() {
		t := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-t.C:
				cond.L.Lock()
				n += 1
				fmt.Println("inc", n)
				cond.Broadcast()
				cond.L.Unlock()
				if n == 10 {
					return
				}
			}
		}
	}()

	wg.Wait()
	fmt.Println("ending", n)
}
```
