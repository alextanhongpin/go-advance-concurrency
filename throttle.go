package main

import (
	"log"
	"sync"
	"time"
)

func main() {

	maxTask := 100
	maxBuff := 5
	numWorkers := 5

	buff := make(chan interface{}, maxBuff)

	// Fill up buffer first
	for i := 0; i < maxBuff; i++ {
		buff <- i
	}
	var wg sync.WaitGroup

	wg.Add(maxTask)
	start := time.Now()

	go func() {
		// Separate processing
		for i := 0; i < maxTask; i++ {
			// Run sequentially
			buff <- i
			log.Println("do something with...", i)
			wg.Done()

			// Run concurrently has no impact if the buffer is filled, since it takes the same amount of time to clear it up
			// go func(i int) {
			// 	buff <- i
			// 	log.Println("do something with...", i)
			// 	wg.Done()
			// }(i)
		}
	}()

	// Run a separate tasks that runs dequeuing every 1 second
	go func() {
		// Rate limiter
		// For example, we want to process 10 items per second
		// so each of them will take 100ms
		// If we have 100 items, it will take 10 seconds

		// To speed up the dequeuing process, we create five different workers that will periodically
		// pull out items from queue
		for i := 0; i < numWorkers; i++ {
			go func(i int) {
				log.Println("worker", i)
				for _ = range time.Tick(100 * time.Millisecond) {
					select {
					case _, ok := <-buff:
						if !ok {
							// Will only reach here if the buff channel is closed
							return
						}
					}
					log.Println("current buffer", len(buff))
				}
			}(i)
		}

		log.Println("throttle complete")
	}()

	wg.Wait()
	close(buff)

	log.Println("done", time.Since(start))
}
