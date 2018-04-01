package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

func main() {
	in := make(chan interface{})

	done := make(chan interface{})
	defer close(done)

	var mu sync.Mutex

	go func() {
		count := 0
		var wg sync.WaitGroup

		wg.Add(5000)
		for i := 0; i < 100; i++ {
			go func() {
				for j := 0; j < 50; j++ {
					mu.Lock()
					count++
					mu.Unlock()

					time.Sleep(time.Duration(rand.Intn(500)+500) * time.Millisecond)
					in <- j
					wg.Done()
				}
			}()
		}
		wg.Wait()
		close(in)
		log.Printf("%d items sent\n", count)
	}()

	var wg sync.WaitGroup
	wg.Add(5000)
	count := 0
	for v := range batchEvent(done, in) {
		log.Printf("done, got %d items\n", len(v))
		count += len(v)
		for _ = range v {
			wg.Done()
		}
	}
	wg.Wait()

	log.Printf("%d items received\n", count)
}

func batchEvent(done, in <-chan interface{}) <-chan []int {
	batchSize := 100
	out := make(chan []int)
	batch := []int{}

	go func() {
		defer close(out)

		for {
			select {
			case <-done:
				return
			case e, ok := <-in:
				if !ok {
					log.Println("no more items, closing out channel")
					// Clear up channel
					if (len(batch)) > 0 {
						out <- batch
						batch = []int{}
					}
					return
				}
				batch = append(batch, e.(int))

				if len(batch) == batchSize {
					log.Println("batch is full, processing")
					out <- batch
					batch = []int{}
				}
			case <-time.After(50 * time.Millisecond):
				if len(batch) > 0 {
					log.Println("time exceeded 100 ms, batching requests")
					out <- batch
					batch = []int{}
				}
			}
		}
	}()

	return out
}

// 2018/03/20 18:26:19 batch is full, processing
// 2018/03/20 18:26:19 done, got 100 items
// 2018/03/20 18:26:19 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:19 done, got 2 items
// 2018/03/20 18:26:20 batch is full, processing
// 2018/03/20 18:26:20 done, got 100 items
// 2018/03/20 18:26:21 batch is full, processing
// 2018/03/20 18:26:21 done, got 100 items
// 2018/03/20 18:26:22 batch is full, processing
// 2018/03/20 18:26:22 done, got 100 items
// 2018/03/20 18:26:22 batch is full, processing
// 2018/03/20 18:26:22 done, got 100 items
// 2018/03/20 18:26:23 batch is full, processing
// 2018/03/20 18:26:23 done, got 100 items
// 2018/03/20 18:26:24 batch is full, processing
// 2018/03/20 18:26:24 done, got 100 items
// 2018/03/20 18:26:25 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:25 done, got 80 items
// 2018/03/20 18:26:25 batch is full, processing
// 2018/03/20 18:26:25 done, got 100 items
// 2018/03/20 18:26:26 batch is full, processing
// 2018/03/20 18:26:26 done, got 100 items
// 2018/03/20 18:26:27 batch is full, processing
// 2018/03/20 18:26:27 done, got 100 items
// 2018/03/20 18:26:28 batch is full, processing
// 2018/03/20 18:26:28 done, got 100 items
// 2018/03/20 18:26:28 batch is full, processing
// 2018/03/20 18:26:28 done, got 100 items
// 2018/03/20 18:26:29 batch is full, processing
// 2018/03/20 18:26:29 done, got 100 items
// 2018/03/20 18:26:30 batch is full, processing
// 2018/03/20 18:26:30 done, got 100 items
// 2018/03/20 18:26:31 batch is full, processing
// 2018/03/20 18:26:31 done, got 100 items
// 2018/03/20 18:26:31 batch is full, processing
// 2018/03/20 18:26:31 done, got 100 items
// 2018/03/20 18:26:32 batch is full, processing
// 2018/03/20 18:26:32 done, got 100 items
// 2018/03/20 18:26:33 batch is full, processing
// 2018/03/20 18:26:33 done, got 100 items
// 2018/03/20 18:26:34 batch is full, processing
// 2018/03/20 18:26:34 done, got 100 items
// 2018/03/20 18:26:34 batch is full, processing
// 2018/03/20 18:26:34 done, got 100 items
// 2018/03/20 18:26:35 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:35 done, got 81 items
// 2018/03/20 18:26:36 batch is full, processing
// 2018/03/20 18:26:36 done, got 100 items
// 2018/03/20 18:26:37 batch is full, processing
// 2018/03/20 18:26:37 done, got 100 items
// 2018/03/20 18:26:37 batch is full, processing
// 2018/03/20 18:26:37 done, got 100 items
// 2018/03/20 18:26:38 batch is full, processing
// 2018/03/20 18:26:38 done, got 100 items
// 2018/03/20 18:26:39 batch is full, processing
// 2018/03/20 18:26:39 done, got 100 items
// 2018/03/20 18:26:40 batch is full, processing
// 2018/03/20 18:26:40 done, got 100 items
// 2018/03/20 18:26:40 batch is full, processing
// 2018/03/20 18:26:40 done, got 100 items
// 2018/03/20 18:26:41 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:41 done, got 41 items
// 2018/03/20 18:26:41 batch is full, processing
// 2018/03/20 18:26:41 done, got 100 items
// 2018/03/20 18:26:42 batch is full, processing
// 2018/03/20 18:26:42 done, got 100 items
// 2018/03/20 18:26:43 batch is full, processing
// 2018/03/20 18:26:43 done, got 100 items
// 2018/03/20 18:26:44 batch is full, processing
// 2018/03/20 18:26:44 done, got 100 items
// 2018/03/20 18:26:44 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:44 done, got 60 items
// 2018/03/20 18:26:45 batch is full, processing
// 2018/03/20 18:26:45 done, got 100 items
// 2018/03/20 18:26:46 batch is full, processing
// 2018/03/20 18:26:46 done, got 100 items
// 2018/03/20 18:26:46 batch is full, processing
// 2018/03/20 18:26:46 done, got 100 items
// 2018/03/20 18:26:47 batch is full, processing
// 2018/03/20 18:26:47 done, got 100 items
// 2018/03/20 18:26:48 batch is full, processing
// 2018/03/20 18:26:48 done, got 100 items
// 2018/03/20 18:26:49 batch is full, processing
// 2018/03/20 18:26:49 done, got 100 items
// 2018/03/20 18:26:49 batch is full, processing
// 2018/03/20 18:26:49 done, got 100 items
// 2018/03/20 18:26:50 batch is full, processing
// 2018/03/20 18:26:50 done, got 100 items
// 2018/03/20 18:26:50 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:50 done, got 48 items
// 2018/03/20 18:26:51 batch is full, processing
// 2018/03/20 18:26:51 done, got 100 items
// 2018/03/20 18:26:52 batch is full, processing
// 2018/03/20 18:26:52 done, got 100 items
// 2018/03/20 18:26:53 batch is full, processing
// 2018/03/20 18:26:53 done, got 100 items
// 2018/03/20 18:26:53 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:53 done, got 98 items
// 2018/03/20 18:26:54 batch is full, processing
// 2018/03/20 18:26:54 done, got 100 items
// 2018/03/20 18:26:55 batch is full, processing
// 2018/03/20 18:26:55 done, got 100 items
// 2018/03/20 18:26:56 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:56 done, got 97 items
// 2018/03/20 18:26:56 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:56 done, got 17 items
// 2018/03/20 18:26:57 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:57 done, got 53 items
// 2018/03/20 18:26:57 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:57 done, got 7 items
// 2018/03/20 18:26:57 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:57 done, got 3 items
// 2018/03/20 18:26:57 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:57 done, got 2 items
// 2018/03/20 18:26:58 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:58 done, got 5 items
// 2018/03/20 18:26:58 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:58 done, got 1 items
// 2018/03/20 18:26:58 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:58 done, got 1 items
// 2018/03/20 18:26:58 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:58 done, got 1 items
// 2018/03/20 18:26:58 time exceeded 100 ms, batching requests
// 2018/03/20 18:26:58 done, got 2 items
// 2018/03/20 18:26:58 5000 items sent
// 2018/03/20 18:26:58 no more items, closing out channel
// 2018/03/20 18:26:58 done, got 1 items
// 2018/03/20 18:26:58 5000 items received
