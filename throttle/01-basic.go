package main

import (
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var m sync.Mutex

func main() {

	keywords := strings.Split("abcdefghijklmnopqrstuvwxyz", "")

	// start holds the start time of each goroutine
	start := make(map[string]time.Time, len(keywords))

	// duration holds the time taken for each goroutine to complete
	duration := make(map[int]time.Duration, len(keywords))

	// doWork is a function that "fan-out" to do some work, but can only spawn
	// limited goroutine based on the number of available CPUs
	doWork := func(
		done <-chan interface{},
		worker chan interface{},
		values ...string,
	) <-chan interface{} {
		log.Println("start pipeline")
		stream := make(chan interface{})

		var wg sync.WaitGroup

		// Indicate the number of expected work to be done
		wg.Add(len(values))

		go func() {
			for k, v := range values {
				// Log the start time...
				log.Println("start worker:", k)
				m.Lock()
				start["start:"+strconv.Itoa(k)] = time.Now()
				m.Unlock()
				// "Push" a value to the "queue"
				// Since it is a buffered channel, it will block once it is filled
				worker <- k
				// Create multiple workers here
				go func(v string) {
					defer wg.Done()
					defer func() {
						// Job's done, "pop" the value from the "queue"
						out := <-worker
						log.Println("stop worker:", out)
						// Calculate the duration taken for each goroutine from the start time
						m.Lock()
						duration[out.(int)+1] = time.Since(start["start:"+strconv.Itoa(out.(int))])
						m.Unlock()
					}()

					// Sleep to mimic heavy processing here

					time.Sleep(500 * time.Millisecond)
					select {
					// You can force the goroutine to end by closing the parent channel
					// This is to prevent goroutine from leaking
					case <-done:
						return
						// Send the value to the stream
					case stream <- v:
					}
				}(v)
			}
		}()

		// Close the channel on completion to signal the end
		go func() {
			// Wait for all the goroutine to complete
			wg.Wait()

			// Close the stream to indicate that all the values have been processed
			close(stream)
			log.Println("stop pipeline")
		}()
		return stream
	}

	// We use the numCPUs as the maximum number of workers available at a time -
	// a.k.a the maximum number of goroutine that can be spawned at a time
	numCPUs := runtime.NumCPU()

	// global parent channel, close this to signal all goroutines to close
	done := make(chan interface{})

	// buffered channel to indicate the maximum number of goroutines that can be spawned
	worker := make(chan interface{}, numCPUs)

	// Close all channels at the end to prevent goroutine leak
	defer close(worker)
	defer close(done)

	t0 := time.Now()
	for v := range doWork(done, worker, keywords...) {
		log.Println("result:", v)
	}
	log.Println("time taken:", time.Since(t0))

	// Log the duration for each goroutine
	var total float64
	for k, v := range keywords {
		_ = v
		log.Printf("worker %v took %v", k+1, duration[k+1])
		total += duration[k+1].Seconds()
	}
	log.Println(total)

	log.Println("process exiting")
}
