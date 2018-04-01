package main

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	numWorkers := 10
	start := time.Now()
	chain := NewChain(ctx, generator(ctx, 100))
	chain.
		Throttle(5, true).
		Pipe(func(in interface{}) interface{} {
			v := in.(int)
			// time.Sleep(time.Duration(rand.Intn(50)+50) * time.Millisecond)
			time.Sleep(100 * time.Millisecond)
			return v * 2
		}).
		// Pool(numWorkers, true, func(in interface{}) interface{} {
		// 	v := in.(int)
		// 	// Simulate slow worker
		// 	time.Sleep(time.Duration(rand.Intn(150)+150) * time.Millisecond)
		// 	return v - 50
		// }).
		Pipe(func(in interface{}) interface{} {
			v := in.(int)
			// Simulate slow output, which will throttle the input
			time.Sleep(50 * time.Millisecond)
			// time.Sleep(time.Duration(rand.Intn(150)+150) * time.Millisecond)
			return v + 10
		}).
		Drain(func(in interface{}) {
			log.Println(in)
		})

	log.Printf("time taken for %d worker %v\n", numWorkers, time.Since(start))
	// Without throttling
	// time taken for 1 worker 22.658940917s
	// time taken for 5 worker 4.667852926s
	// time taken for 10 worker 2.331470883s
}

type GenericFunc func(interface{}) interface{}

type Stream interface {
	Pipe(GenericFunc) Stream
	Throttle() Stream
	Pool() Stream
	Drain(func(interface{}))
}

type Chain struct {
	Ctx   context.Context
	Queue chan interface{}
}

func (c *Chain) Pipe(fn GenericFunc) *Chain {
	outStream := make(chan interface{})

	go func() {
		defer close(outStream)
		for i := range c.Queue {
			select {
			case <-c.Ctx.Done():
				return
			case outStream <- fn(i):
			}
		}
	}()

	return &Chain{
		Ctx:   c.Ctx,
		Queue: outStream,
	}
}

func (c *Chain) Throttle(threshold int, debug bool) *Chain {
	if threshold == 0 {
		threshold = 10
	}
	outStream := make(chan interface{}, threshold)

	go func() {
		defer close(outStream)
		for i := range c.Queue {
			select {
			case <-c.Ctx.Done():
				return
			case outStream <- i:
				if debug {
					log.Println("capacity:", len(outStream))
				}
			}
		}
	}()

	return &Chain{
		Ctx:   c.Ctx,
		Queue: outStream,
	}
}

func (c *Chain) Pool(numWorkers int, debug bool, fn GenericFunc) *Chain {
	if numWorkers == 0 {
		numWorkers = 10
	}
	outStream := make(chan interface{})

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	worker := func(index int, in chan interface{}) {
		defer wg.Done()
		for i := range in {
			select {
			case <-c.Ctx.Done():
				return
			case outStream <- fn(i):
				log.Println("worker", index)
			}
		}
	}

	for i := 0; i < numWorkers; i++ {
		go worker(i, c.Queue)
	}

	go func() {
		wg.Wait()
		close(outStream)
	}()

	return &Chain{
		Ctx:   c.Ctx,
		Queue: outStream,
	}
}

func (c *Chain) Drain(fn func(interface{})) {
	for i := range c.Queue {
		fn(i)
	}
}

func NewChain(ctx context.Context, in chan interface{}) *Chain {
	return &Chain{
		Ctx:   ctx,
		Queue: in,
	}
}

func generator(ctx context.Context, limit int) chan interface{} {
	outStream := make(chan interface{})

	go func() {
		defer close(outStream)
		for i := 0; i < limit; i++ {
			select {
			case <-ctx.Done():
				return
			case outStream <- i:
			}
		}
	}()

	return outStream
}
