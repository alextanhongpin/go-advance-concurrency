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
