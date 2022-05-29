# Run a fixed number of worker at a time

```go
// You can edit this code!
// Click here and start typing.
package main

import (
	"fmt"
	"sync"
)

func main() {
	items := 20
	workers := 4

	ch := make(chan int)

	go func() {
		for i := 0; i < items; i++ {
			ch <- i
		}
		close(ch)
	}()

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()

			for {
				select {
				case n, open := <-ch:
					if !open {
						fmt.Println("worker closing", i)
						return
					}
					fmt.Println("worker", i, n)
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("exiting")
}
```

## Run at most n goroutine

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
	items := 20
	workers := 4

	ch := make(chan int)

	go func() {
		for i := 0; i < items; i++ {
			ch <- i
		}
		close(ch)
	}()

	var wg sync.WaitGroup
	sem := make(chan struct{}, workers)

	for n := range ch {
		sem <- struct{}{}
		wg.Add(1)
		fmt.Println("start", n)
		go func(n int) {
			defer func() {
				fmt.Println("done", n)
				wg.Done()
				<-sem
			}()

			fmt.Println("working on", n)
			time.Sleep(100 * time.Millisecond)
		}(n)
	}

	wg.Wait()
	fmt.Println("exiting")
}
```
