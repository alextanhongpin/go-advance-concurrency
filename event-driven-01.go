// This program demonstrates the event-driven processing concepts found in the
// book Designing Distributed System

package main

import (
	"context"
	"log"
	"sync"
)

func main() {
	generator := func(ctx context.Context, limit int) <-chan interface{} {
		outStream := make(chan interface{})

		go func() {
			for i := 0; i < limit; i++ {
				select {
				case <-ctx.Done():
					return
				case outStream <- i:
				}
			}
			// Preferable over `defer close(outStream)` due to the overhead
			// of calling defer
			close(outStream)
		}()

		return outStream
	}

	copier := func(ctx context.Context, inStream <-chan interface{}) (<-chan interface{}, <-chan interface{}) {
		outStream1 := make(chan interface{})
		outStream2 := make(chan interface{})

		go func() {
			for i := range inStream {
				select {
				case <-ctx.Done():
					return
				case outStream1 <- i:
				}

				select {
				case <-ctx.Done():
					return
				case outStream2 <- i:
				}
			}

			close(outStream1)
			close(outStream2)
		}()

		return outStream1, outStream2
	}

	filter := func(ctx context.Context, inStream <-chan interface{}, fn func(interface{}) bool) <-chan interface{} {
		outStream := make(chan interface{})

		go func() {
			for i := range inStream {
				if !fn(i) {
					continue
				}
				select {
				case <-ctx.Done():
					return
				case outStream <- i:
				}
			}
			close(outStream)
		}()

		return outStream
	}

	splitter := func(ctx context.Context, inStream <-chan interface{}, fn func(interface{}) bool) (<-chan interface{}, <-chan interface{}) {
		outStream1 := make(chan interface{})
		outStream2 := make(chan interface{})

		go func() {
			for i := range inStream {

				if fn(i) {
					select {
					case <-ctx.Done():
						return
					case outStream1 <- i:
					}
				} else {
					select {
					case <-ctx.Done():
						return
					case outStream2 <- i:
					}
				}
			}

			close(outStream1)
			close(outStream2)
		}()

		return outStream1, outStream2
	}

	sharder := func(ctx context.Context, inStream <-chan interface{}, fn func(interface{}) interface{}, numWorkers int) <-chan interface{} {
		outStream := make(chan interface{})

		var wg sync.WaitGroup
		wg.Add(numWorkers)

		worker := func(index int, in <-chan interface{}) {

			for i := range in {
				select {
				case <-ctx.Done():
					return
				case outStream <- fn(i):
					// log.Println("worker ", index, i)
				}
			}
			wg.Done()
		}

		for i := 0; i < numWorkers; i++ {
			go worker(i, inStream)
		}

		go func() {
			wg.Wait()
			close(outStream)
		}()

		return outStream
	}

	merger := func(ctx context.Context, streams ...<-chan interface{}) <-chan interface{} {
		outStream := make(chan interface{})

		var wg sync.WaitGroup
		wg.Add(len(streams))

		worker := func(in <-chan interface{}) {
			for i := range in {
				select {
				case <-ctx.Done():
					return
				case outStream <- i:
				}
			}
			wg.Done()
		}

		for _, i := range streams {
			go worker(i)
		}

		go func() {
			wg.Wait()
			close(outStream)
		}()

		return outStream
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	numItems := 100

	outStream1, outStream2 := copier(ctx, generator(ctx, numItems))

	evenStream := filter(ctx, outStream1, even)
	oddStream := filter(ctx, outStream2, odd)

	greaterThanNumStream, lessThanNumStream := splitter(ctx, evenStream, greaterThan(numItems/2))

	doubleOddStream := sharder(ctx, oddStream, double, 4)

	out := merger(ctx, greaterThanNumStream, lessThanNumStream, doubleOddStream)

	for o := range out {
		log.Println(o)
	}
	log.Println("done")
}

func even(i interface{}) bool {
	v := i.(int)
	return v%2 == 0
}

func odd(i interface{}) bool {
	v := i.(int)
	return v%2 != 0
}

func double(i interface{}) interface{} {
	v := i.(int)
	return v * 2
}

func greaterThan(n int) func(interface{}) bool {
	return func(i interface{}) bool {
		v := i.(int)
		return v > n
	}
}
