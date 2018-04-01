// This program demonstrates how to pause all running goroutines when an error
// occured in one of them. Consider this as a global circuit-breaker

package main

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"
)

func main() {
	numWorkers := 5

	var mu sync.Mutex
	cond := sync.NewCond(&mu)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var pause bool

	counter := 0
	start := time.Now()
	for i := range worker(ctx, generator(ctx, 100), numWorkers, cond, pause) {
		counter++
		log.Println(i)
	}

	log.Println(time.Since(start))
	log.Println("done", counter)
}

func generator(ctx context.Context, limit int) <-chan interface{} {
	outStream := make(chan interface{})

	go func() {
		defer close(outStream)
		for i := 0; i < limit; i++ {
			select {
			// The sequence of context done, does it matters?
			case outStream <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	return outStream
}

func worker(ctx context.Context, inStream <-chan interface{}, numWorkers int, cond *sync.Cond, pause bool) <-chan interface{} {

	outStream := make(chan interface{})

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	failIndex := rand.Intn(50)
	multiplex := func(index int, in <-chan interface{}) {
		log.Println("starting multiplex", index)
		for i := range in {
			// Assuming each operation takes roughly 100 ms
			time.Sleep(time.Duration(rand.Intn(50)+50) * time.Millisecond)

			// Fake failures
			if i == failIndex {
				cond.L.Lock()
				// Set pause to true
				pause = i == failIndex
				log.Println("worker", index, "encountered error at", i)
				cond.L.Unlock()

				// Retry after 5 seconds
				go func() {
					log.Println("spending five seconds to recover")
					time.Sleep(5 * time.Second)

					// Reset (do I need to lock here too?)
					pause = false
					cond.Broadcast()
				}()
			}

			cond.L.Lock()
			// While pause is true...
			for pause {
				// This does not consume cpu cycle. It suspends the current goroutines.
				log.Println("worker", index, "is paused")
				cond.Wait()
			}
			cond.L.Unlock()

			select {
			case <-ctx.Done():
				return
			case outStream <- i:
			}
		}
		wg.Done()
	}

	for i := 0; i < numWorkers; i++ {
		log.Println("executing workers", i)
		go multiplex(i, inStream)
	}

	go func() {
		wg.Wait()
		close(outStream)
	}()

	return outStream
}
