package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

func main() {
	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()

	maxTask := 100
	maxBuff := 5
	numWorkers := 5

	bufferCh := make(chan interface{}, maxBuff)

	// Fill up buffer first. We will periodically clear the buffer to control the
	// rate of processing.
	for i := 0; i < maxBuff; i++ {
		bufferCh <- i
	}

	var wg sync.WaitGroup
	wg.Add(maxTask)

	go func() {
		// Separate processing
		for i := 0; i < maxTask; i++ {
			defer wg.Done()
			// Run sequentially
			bufferCh <- i
			log.Println("processing...", i)

			// Run concurrently has no impact if the buffer is filled, since it takes the same amount of time to clear it up
			// go func(i int) {
			// 	bufferCh <- i
			// 	log.Println("do something with...", i)
			// 	wg.Done()
			// }(i)
		}
	}()

	// Rate limiter
	// Deque the channel every 100 ms.
	// The rate is 10 items per second.

	// To speed up the dequeuing process, we create n different workers that
	// will periodically pull out items from queue.
	//
	// The total duration will then be (number of items) * 100 ms / number of workers.
	// In this example, 100 * 100 ms / 5 = 2000 ms = 2 second.
	deque := func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_, open := <-bufferCh
				if !open {
					// Buffer is closed.
					return
				}
			}
		}
	}

	for i := 0; i < numWorkers; i++ {
		go deque()
	}

	wg.Wait()
	close(bufferCh)

	fmt.Println("program terminating")
}
