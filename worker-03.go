package main

import (
	"context"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// ctx for cancellation :P
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Basically the maximum number of items that we allow in flight
	maxConcurrency := 5

	// The number of workers we want to run concurrently
	maxWorkers := 3

	urls := strings.Split("abcdefghijklmnopqrstuvwxyz", "")

	var wg sync.WaitGroup
	wg.Add(len(urls))
	go func() {
		// Chain the fan-out with throttle
		for v := range throttle(ctx, maxConcurrency, fanIn(ctx, maxWorkers, generator(ctx, urls...))) {
			defer wg.Done()
			// Intentionally delay to view the throttling in action
			time.Sleep(time.Duration(rand.Intn(250)+50) * time.Millisecond)
			log.Println(v)
		}
	}()
	wg.Wait()
	log.Println("done")
}

func generator(ctx context.Context, in ...string) <-chan interface{} {
	log.Println(in)
	outStream := make(chan interface{})

	go func() {
		defer close(outStream)
		for _, v := range in {
			select {
			// Ensure there are no goroutines leak
			case <-ctx.Done():
				return
			case outStream <- v:
			}
		}
	}()

	return outStream
}

func throttle(ctx context.Context, maxConcurrency int, in <-chan interface{}) <-chan interface{} {
	outStream := make(chan interface{}, maxConcurrency)

	go func() {
		defer close(outStream)
		for v := range in {
			select {
			case <-ctx.Done():
				return
			case outStream <- v:
				// Print out the current len and capacity of the queue.
				// The channel will block if the len is equal capacity.
				log.Printf("throttle queue: %d/%d\n", len(outStream), cap(outStream))
			}
		}
	}()

	return outStream
}

func fanIn(ctx context.Context, maxWorkers int, in <-chan interface{}) <-chan interface{} {
	// NOTE: We can add the buffered channels here, but it is always better to decouple logic
	outStream := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(maxWorkers)

	// Multiplex is just a function that processes each channels separately,
	// and sending them back to a single out channel
	multiplex := func(worker int, inStream <-chan interface{}) {
		defer wg.Done()

		for i := range inStream {
			// Fake processing time
			time.Sleep(time.Duration(rand.Intn(250)+250) * time.Millisecond)
			select {
			case <-ctx.Done():
				return
			case outStream <- i:
				log.Println("worker:", worker)
				// log.Printf("fanout queue: %d/%d\n", len(outStream), cap(outStream))
			}
		}
	}

	// For each channel, run a separate process simulating
	// multiple workers
	for i := 0; i < maxWorkers; i++ {
		go multiplex(i, in)
	}

	go func() {
		// Wait for all the workers to be completed before closing the stream
		defer close(outStream)
		wg.Wait()
	}()

	return outStream
}

// go run -race worker-03.go

// 2018/03/27 02:03:40 [a b c d e f g h i j k l m n o p q r s t u v w x y z]
// 2018/03/27 02:03:40 worker 2 c
// 2018/03/27 02:03:40 throttle queue: 0/5
// 2018/03/27 02:03:40 worker 1 b
// 2018/03/27 02:03:40 throttle queue: 1/5
// 2018/03/27 02:03:40 throttle queue: 2/5
// 2018/03/27 02:03:40 worker 0 a
// 2018/03/27 02:03:40 c
// 2018/03/27 02:03:40 worker 2 d
// 2018/03/27 02:03:40 throttle queue: 2/5
// 2018/03/27 02:03:41 worker 0 f
// 2018/03/27 02:03:41 throttle queue: 3/5
// 2018/03/27 02:03:41 worker 1 e
// 2018/03/27 02:03:41 throttle queue: 4/5
// 2018/03/27 02:03:41 b
// 2018/03/27 02:03:41 a
// 2018/03/27 02:03:41 d
// 2018/03/27 02:03:41 worker 2 g
// 2018/03/27 02:03:41 throttle queue: 2/5
// 2018/03/27 02:03:41 worker 0 h
// 2018/03/27 02:03:41 throttle queue: 3/5
// 2018/03/27 02:03:41 worker 1 i
// 2018/03/27 02:03:41 throttle queue: 4/5
// 2018/03/27 02:03:41 f
// 2018/03/27 02:03:41 e
// 2018/03/27 02:03:41 worker 1 l
// 2018/03/27 02:03:41 throttle queue: 3/5
// 2018/03/27 02:03:41 g
// 2018/03/27 02:03:41 worker 0 k
// 2018/03/27 02:03:41 throttle queue: 3/5
// 2018/03/27 02:03:41 throttle queue: 4/5
// 2018/03/27 02:03:41 worker 2 j
// 2018/03/27 02:03:41 h
// 2018/03/27 02:03:42 i
// 2018/03/27 02:03:42 worker 1 m
// 2018/03/27 02:03:42 throttle queue: 3/5
// 2018/03/27 02:03:42 l
// 2018/03/27 02:03:42 worker 0 n
// 2018/03/27 02:03:42 throttle queue: 3/5
// 2018/03/27 02:03:42 k
// 2018/03/27 02:03:42 worker 2 o
// 2018/03/27 02:03:42 throttle queue: 3/5
// 2018/03/27 02:03:42 worker 1 p
// 2018/03/27 02:03:42 throttle queue: 4/5
// 2018/03/27 02:03:42 worker 0 q
// 2018/03/27 02:03:42 throttle queue: 5/5
// 2018/03/27 02:03:42 j
// 2018/03/27 02:03:42 m
// 2018/03/27 02:03:42 worker 2 r
// 2018/03/27 02:03:42 throttle queue: 4/5
// 2018/03/27 02:03:42 worker 0 t
// 2018/03/27 02:03:42 throttle queue: 5/5
// 2018/03/27 02:03:42 worker 1 s
// 2018/03/27 02:03:42 n
// 2018/03/27 02:03:42 throttle queue: 5/5
// 2018/03/27 02:03:43 o
// 2018/03/27 02:03:43 worker 2 u
// 2018/03/27 02:03:43 throttle queue: 5/5
// 2018/03/27 02:03:43 worker 0 v
// 2018/03/27 02:03:43 p
// 2018/03/27 02:03:43 throttle queue: 5/5
// 2018/03/27 02:03:43 worker 1 w
// 2018/03/27 02:03:43 q
// 2018/03/27 02:03:43 throttle queue: 5/5
// 2018/03/27 02:03:43 worker 1 z
// 2018/03/27 02:03:43 r
// 2018/03/27 02:03:43 throttle queue: 5/5
// 2018/03/27 02:03:43 worker 0 y
// 2018/03/27 02:03:43 t
// 2018/03/27 02:03:43 throttle queue: 5/5
// 2018/03/27 02:03:43 worker 2 x
// 2018/03/27 02:03:43 s
// 2018/03/27 02:03:43 throttle queue: 5/5
// 2018/03/27 02:03:44 u
// 2018/03/27 02:03:44 v
// 2018/03/27 02:03:44 w
// 2018/03/27 02:03:44 z
// 2018/03/27 02:03:44 y
// 2018/03/27 02:03:45 x
// 2018/03/27 02:03:45 done
