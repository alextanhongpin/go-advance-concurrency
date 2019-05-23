A pattern to gracefully shutting down goroutines before exiting the program.
```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	close := hoc()
	close()
	close()
	close()
	fmt.Println("Hello, playground")
}

func hoc() func() {
	// Channel to signal completion.
	done := make(chan struct{})

	// Ensure the channels are closed gracefully before the program terminates.
	var wg sync.WaitGroup
	wg.Add(1)

	// Ensure the channels are closed only once. Closing twice will panic.
	var once sync.Once

	// Fake work.
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				fmt.Println("closing job")
				return
			case <-t.C:
				fmt.Println("running background job")
			}
		}
	}()
	return func() {
		fmt.Println("calling close")
		// This will only be called once.
		once.Do(func() {
			fmt.Println("initiate closing")
			close(done)
			wg.Wait()
			fmt.Println("completed")
		})
	}
}
```
