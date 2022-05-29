package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()

	// Set the max concurrency to 5
	maxConcurrency := 3

	// Use a buffered channel to simulate semaphore.
	sem := make(chan struct{}, maxConcurrency)

	nTasks := 10

	var wg sync.WaitGroup

	for i := 0; i < nTasks; i++ {
		// Acquire the semaphore. This will block once it is full.
		// This will prevent spawning too many goroutines.
		sem <- struct{}{}
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			defer func() {
				// Release the semaphore once it is done.
				<-sem

				fmt.Println("done work", i)
			}()

			fmt.Println("performing work:", i)
			time.Sleep(1 * time.Second)
		}(i)
	}

	wg.Wait()

	fmt.Println("program terminating")
}
