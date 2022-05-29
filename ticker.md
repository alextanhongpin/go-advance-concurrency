# Periodic task with ticker

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	every := 100 * time.Millisecond

	t := time.NewTicker(every)
	defer t.Stop()

	done := make(chan interface{})
	go func() {
		time.Sleep(1 * time.Second)
		close(done)
	}()

	for {
		select {
		case <-done:
			return
		case t0 := <-t.C:
			fmt.Println("ticking", t0)
		}
	}
}
```

## Run task every n items or every t duration

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	maxKeys := 5

	var wg sync.WaitGroup
	wg.Add(2)

	ch := make(chan int)
	go func() {
		defer wg.Done()

		for i := 0; i < 3; i++ {
			time.Sleep(100 * time.Millisecond)
			ch <- i
		}

		time.Sleep(1 * time.Second)

		for i := 3; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			ch <- i
		}

		close(ch)
	}()

	go func() {
		defer wg.Done()

		every := 500 * time.Millisecond

		t := time.NewTicker(every)
		defer t.Stop()

		// Run every time we have 5 keys or after 500 milliseconds.
		keys := make([]int, 0, 10)

		for {
			select {
			case <-t.C:
				flush(keys)
				fmt.Println("timer exceeded", keys)
				keys = nil
			case n, open := <-ch:
				if !open {
					// NOTE: Remember to clear all pending keys
					flush(keys)
					fmt.Println("flush remaining", keys)
					keys = nil
					return
				}
				fmt.Println("received", n)

				// Perform a debounce every time a key is received.
				t.Reset(every)

				keys = append(keys, n)

				// If maxKeys is not set, then only clear the keys in every tick duration.
				if maxKeys == 0 || len(keys) < maxKeys {
					continue
				}
				flush(keys)
				fmt.Println("threshold exceeded", keys)
				keys = nil
			}
		}
	}()
	wg.Wait()

	fmt.Println("program terminating")
}

func flush(keys []int) {
	if len(keys) == 0 {
		return
	}
}
```
