package main

import (
	"context"
	"log"
	"sync"
)

func main() {

	generator := func(n int) <-chan interface{} {
		stream := make(chan interface{})

		go func() {
			defer close(stream)
			for i := 0; i < n; i++ {
				stream <- i
			}
		}()

		return stream
	}

	split := func(ctx context.Context, in <-chan interface{}) (<-chan interface{}, <-chan interface{}) {
		out1 := make(chan interface{})
		out2 := make(chan interface{})

		go func() {
			defer close(out1)
			defer close(out2)
			for i := range in {
				select {
				case out1 <- i:
				case <-ctx.Done():
					return
				}

				select {
				case out2 <- i:
				case <-ctx.Done():
					return
				}
			}
		}()

		return out1, out2
	}

	var wg sync.WaitGroup
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(2)

	o1, o2 := split(ctx, generator(10))

	go func() {
		defer wg.Done()
		for o := range o1 {
			log.Println("split 1:", o)
		}
	}()

	go func() {
		defer wg.Done()
		for o := range o2 {
			log.Println("split 2:", o)
		}
	}()

	wg.Wait()
	log.Println("terminating program")
}
