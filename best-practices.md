# Best practices


## Putting wg.Done in task definition

The `listen` method does not need to know about `wg.Done`.


```go
	var wg sync.WaitGroup
	wg.Add(n)

	// Bad practice: putting wg.Done in task definition
	listen := func(i int) {
		defer wg.Done()

		for msg := range ch {
			fmt.Printf("consumer#%d: %d\n", i, msg)
		}
	}
	for i := 0; i < n; i++ {
		go func(i int) {
			listen(i)
		}(i)
	}
```

A better approach is to split the function into two
- `listen` which is synchronous
- `listenAsync` which is concurrent

```go
	var wg sync.WaitGroup

	// Bad practice: putting wg.Done in task definition
	listen := func(i int) {
		for msg := range ch {
			fmt.Printf("consumer#%d: %d\n", i, msg)
		}
	}
	listenAsync := func(i int) {
		wg.Add(1)
		go func() {
			defer wg.Done()

			listen(i)
		}()
	}

	for i := 0; i < n; i++ {
		listenAsync(i)
	}
```

## Not limiting concurrent goroutine

The example below works fine if the number of tasks is known and is small. However, if the number of tasks is large, e.g. 1 million, you might end up creating a milliion goroutine.

```go
// You can edit this code!
// Click here and start typing.
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()
	tasks := 10
	var wg sync.WaitGroup
	wg.Add(tasks)

	for i := 0; i < tasks; i++ {
		go func(i int) {
			defer wg.Done()

			duration := randomDuration(100*time.Millisecond, 1*time.Second)
			fmt.Println("do task", i, duration)
			time.Sleep(duration)
		}(i)
	}

	wg.Wait()
}

func randomDuration(min, max time.Duration) time.Duration {
	lo := min.Milliseconds()
	hi := max.Milliseconds()
	duration := time.Duration(rand.Int63n(hi-lo)+lo) * time.Millisecond
	return duration
}
```

We can solve this by using a buffered channel to simulate a semaphore:

```go
// You can edit this code!
// Click here and start typing.
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()
	tasks := 10
	var wg sync.WaitGroup
	wg.Add(tasks)

	maxGoroutine := 5
	semaphore := make(chan struct{}, maxGoroutine)

	for i := 0; i < tasks; i++ {
		// This will start blocking once it is full.
		// Once the task in the goroutine is completed,
		// then a new goroutine can be started.
		semaphore <- struct{}{}

		go func(i int) {
			defer wg.Done()
			defer func() {
				<-semaphore
			}()

			duration := randomDuration(1*time.Second, 2*time.Second)
			fmt.Println("do task", i, duration)
			time.Sleep(duration)
		}(i)
	}

	wg.Wait()
}

func randomDuration(min, max time.Duration) time.Duration {
	lo := min.Milliseconds()
	hi := max.Milliseconds()
	duration := time.Duration(rand.Int63n(hi-lo)+lo) * time.Millisecond
	return duration
}
```

## Not having proper initialization/closing of goroutine

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	bg := NewBackground()
	// Someone can forget to call this, causing goroutine leak.
	// This might be called multiple times too, calling close on closed channel.
	defer bg.Stop()

	// Someone may forget to call this too.
	// This might be called multiple times too, spawning more background
	// goroutine (which may be unintended)
	bg.Start()

	// Send may happen after channel is closed too.
	bg.Send(1)
	fmt.Println("program terminating")
}

type Background struct {
	quit chan struct{}
	wg   sync.WaitGroup
	ch   chan int
}

func NewBackground() *Background {
	return &Background{
		quit: make(chan struct{}),
		ch:   make(chan int),
	}
}

func (b *Background) Start() {
	b.wg.Add(1)

	go func() {
		defer b.wg.Done()

		for {
			select {
			case <-b.quit:
				fmt.Println("quitting")
				return
			case n := <-b.ch:
				fmt.Println("received", n)
			}
		}
	}()
}

func (b *Background) Stop() {
	close(b.quit)
	b.wg.Wait()
}

func (b *Background) Send(n int) {
	b.ch <- n
}
```


How to improve
```go
// You can edit this code!
// Click here and start typing.
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	bg, stop := NewBackground(5)

	go func() {
		time.Sleep(500 * time.Millisecond)
		stop()
	}()
	n := 10
	for i := 0; i < n; i++ {
		go func(i int) {
			time.Sleep(100 * time.Millisecond)
			bg.Send(i)
		}(i)
	}
	select {
	case <-time.After(5 * time.Second):
		fmt.Println("program terminating")
	}
}

type Background struct {
	wg   sync.WaitGroup
	ch   chan int
	init sync.Once
	quit chan struct{}
}

func NewBackground(n int) (*Background, func()) {
	b := &Background{
		quit: make(chan struct{}),
		ch:   make(chan int, n),
	}

	var once sync.Once
	return b, func() {
		b.init.Do(func() {
			// In case stop is called before Send, then
			// this will prevent the goroutine from being spawned.
		})
		// Ensure this is closed once.
		once.Do(func() {
			close(b.quit)
			if len(b.ch) > 0 {
				close(b.ch)
			}
			b.wg.Wait()
			fmt.Println("done")
		})
	}
}

/* Separate the loop and loopAsync. The loopAsync is meant for coordination, and hence increments/decrements the wait group.
This makes it easier when we need to scale the number of background workers too, e.g.

func (b *Background) loopAsync() {
	b.wg.Add(b.maxWorker)

	for i := 0; i < b.maxWorker; i++ {
		go func() {
			defer b.wg.Done()
			b.loop()
		}()
	}
}
*/
func (b *Background) loopAsync() {
	b.wg.Add(1)

	go func() {
		defer b.wg.Done()

		b.loop()
		fmt.Println("closing loop")
	}()
}

func (b *Background) loop() {
	for {
		select {
		case <-b.quit:
			fmt.Println("quitting")
			// Uncomment this when using buffered channel to clear the remaining items
			for len(b.ch) > 0 {
				fmt.Println("flushing", <-b.ch)
			}
			return
		case n := <-b.ch:
			fmt.Println("received", n)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (b *Background) Send(n int) bool {
	b.init.Do(func() {
		// Lazily start the goroutine only when send is first invoked.
		// This prevents creating too many goroutine when it is not used.
		b.loopAsync()
	})

	select {
	// Don't send when the channel is closed.
	case <-b.quit:
		return false
	case b.ch <- n:
		fmt.Println("send", n)
		return true
	}
}
```
