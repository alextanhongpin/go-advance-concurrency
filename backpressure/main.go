package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	// Increase the number of worker to process more items.
	numWorker := 1

	// Create a queue with max buffer of 10 items.
	ch := make(chan int, 10)

	// Our SLA is 10 millisecond of processing time.
	// If the queue is full, and the SLA is violated, we drop the event and notify the user.
	sla := 25 * time.Millisecond

	send := func(n int) bool {
		select {
		case ch <- n:
			return true
		case <-time.After(sla):
			return false
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorker; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			var count int
			// Simulate a slow consumer
			for range ch {
				sleep := time.Duration(rand.Intn(50)) * time.Millisecond
				time.Sleep(sleep)
				count++
			}

			fmt.Println("recv:", count, "worker:", i)
		}(i)
	}

	n := 50
	fmt.Println("send:", n)
	for i := 0; i < n; i++ {
		// TODO: Notify user that their request will not be handled.
		if !send(i) {
			fmt.Println("drop:", i)
		}
	}
	close(ch)
	wg.Wait()

	fmt.Println("program exiting")
}
