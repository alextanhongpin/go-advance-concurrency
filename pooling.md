# Pooling

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"sync/atomic"
)

func main() {
	// Pattern: Lazy pooler.
	// A pooler that runs at every interval, but with increasing interval if
	// there are no jobs to be processed.

	ch := make(chan int)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		time.Sleep(7 * time.Second)
		ch <- 10
	}()

	go func() {
		time.Sleep(15 * time.Second)
		fmt.Println("signalling closing")
		cancel()
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		var i atomic.Int64

		t := time.NewTicker(1 * time.Second)
		defer t.Stop()

		now := time.Now()

		for {
			select {
			case <-t.C:

				// At every tick, check if there are jobs to be processed in the db.
				// If no, then increment the ticker duration.
				var hasResult bool
				if !hasResult {
					switch n := i.Add(1); n {
					case 5:
						fmt.Println("set to 2 second now")
						t.Reset(2 * time.Second)
					case 10:
						t.Reset(10 * time.Second)
					case 20:
						t.Reset(1 * time.Minute)
					default:
						// Prevent overflow?
						if n > 1_000_000 {
							i.Store(0)
						}
					}
				} else {
					// Reset the ticker.
					n := i.Swap(0)
					if n > 0 {
						t.Reset(1 * time.Second)
					}
				}

				fmt.Println("ticking", time.Since(now))
			case v := <-ch:
				// We can receive signal to reset the timer too.
				old := i.Swap(0)
				if old > 0 {
					fmt.Println("resetting")
					t.Reset(1 * time.Second)
				}

				fmt.Println("v", v)
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
	fmt.Println("exiting")
}
```
