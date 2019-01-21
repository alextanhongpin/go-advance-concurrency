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
