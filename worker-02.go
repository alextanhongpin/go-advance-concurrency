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

	// For each worker, find out the number of items they can process
	batch := int(len(urls)/maxWorkers) + 1

	// Create worker channels to store their job
	workers := make([]<-chan interface{}, maxWorkers)

	// NOTE: See the worker-03.go example, this step is unnecessary
	// For each worker...
	for i := 0; i < maxWorkers; i++ {
		// Get the start index
		start := i * batch

		// ...and the end index
		end := (i + 1) * batch

		// Make sure that the index is within slice range...
		if i == maxWorkers-1 {
			end = len(urls)
		}
		// Split urls into separate sections to be processed separately
		// and convert each items into a channel stream
		workers[i] = generator(ctx, urls[start:end]...)
	}

	var wg sync.WaitGroup
	wg.Add(len(urls))
	go func() {
		// Chain the fan-out with throttle
		for v := range throttle(ctx, maxConcurrency, fanIn(ctx, workers...)) {
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

func fanIn(ctx context.Context, in ...<-chan interface{}) <-chan interface{} {
	outStream := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(len(in))

	// Multiplex is just a function that processes each channels separately,
	// and sending them back to a single out channel
	multiplex := func(inStream <-chan interface{}) {
		defer wg.Done()

		for i := range inStream {
			// Fake processing time
			time.Sleep(time.Duration(rand.Intn(250)+250) * time.Millisecond)
			select {
			case <-ctx.Done():
				return
			case outStream <- i:
				// log.Printf("fanout queue: %d/%d\n", len(outStream), cap(outStream))
			}
		}
	}

	// For each channel, run a separate process simulating
	// multiple workers
	for _, i := range in {
		go multiplex(i)
	}

	go func() {
		// Wait for all the workers to be completed before closing the stream
		defer close(outStream)
		wg.Wait()
	}()

	return outStream
}

// go run -race worker-02.go
// 2018/03/26 22:42:21 [a b c d e f g h i]
// 2018/03/26 22:42:21 [j k l m n o p q r]
// 2018/03/26 22:42:21 [s t u v w x y z]
// 2018/03/26 22:42:21 throttle queue: 0/5
// 2018/03/26 22:42:22 throttle queue: 1/5
// 2018/03/26 22:42:22 a
// 2018/03/26 22:42:22 throttle queue: 2/5
// 2018/03/26 22:42:22 j
// 2018/03/26 22:42:22 throttle queue: 1/5
// 2018/03/26 22:42:22 throttle queue: 2/5
// 2018/03/26 22:42:22 throttle queue: 3/5
// 2018/03/26 22:42:22 s
// 2018/03/26 22:42:22 k
// 2018/03/26 22:42:22 throttle queue: 2/5
// 2018/03/26 22:42:22 b
// 2018/03/26 22:42:22 throttle queue: 2/5
// 2018/03/26 22:42:23 t
// 2018/03/26 22:42:23 throttle queue: 2/5
// 2018/03/26 22:42:23 l
// 2018/03/26 22:42:23 throttle queue: 2/5
// 2018/03/26 22:42:23 throttle queue: 3/5
// 2018/03/26 22:42:23 throttle queue: 4/5
// 2018/03/26 22:42:23 c
// 2018/03/26 22:42:23 throttle queue: 4/5
// 2018/03/26 22:42:23 u
// 2018/03/26 22:42:23 throttle queue: 4/5
// 2018/03/26 22:42:23 throttle queue: 5/5
// 2018/03/26 22:42:23 d
// 2018/03/26 22:42:23 throttle queue: 5/5
// 2018/03/26 22:42:23 m
// 2018/03/26 22:42:24 throttle queue: 5/5
// 2018/03/26 22:42:24 v
// 2018/03/26 22:42:24 throttle queue: 5/5
// 2018/03/26 22:42:24 e
// 2018/03/26 22:42:24 throttle queue: 5/5
// 2018/03/26 22:42:24 w
// 2018/03/26 22:42:24 throttle queue: 5/5
// 2018/03/26 22:42:24 n
// 2018/03/26 22:42:24 throttle queue: 5/5
// 2018/03/26 22:42:24 f
// 2018/03/26 22:42:24 throttle queue: 5/5
// 2018/03/26 22:42:24 x
// 2018/03/26 22:42:24 throttle queue: 5/5
// 2018/03/26 22:42:25 o
// 2018/03/26 22:42:25 throttle queue: 5/5
// 2018/03/26 22:42:25 g
// 2018/03/26 22:42:25 throttle queue: 5/5
// 2018/03/26 22:42:25 y
// 2018/03/26 22:42:25 throttle queue: 5/5
// 2018/03/26 22:42:25 p
// 2018/03/26 22:42:26 h
// 2018/03/26 22:42:26 z
// 2018/03/26 22:42:26 q
// 2018/03/26 22:42:26 i
// 2018/03/26 22:42:26 r
// 2018/03/26 22:42:26 done
