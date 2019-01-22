# Fan-in and Fan-out operation in golang

A prime number finder example.

```go
package main

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type PrimeFinder func(done chan interface{}, in <-chan interface{}) <-chan interface{}

func main() {
	randFn := func() interface{} {
		return rand.Int()
	}

	isPrime := func(value int) bool {
		for i := 2; i <= int(math.Floor(math.Sqrt(float64(value)))); i++ {
			if value%i == 0 {
				return false
			}
		}
		return value > 1
	}

	var primeFinder PrimeFinder
	primeFinder = func(done chan interface{}, in <-chan interface{}) <-chan interface{} {
		outStream := make(chan interface{})
		go func() {
			defer close(outStream)
			for {

				select {
				case <-done:
					return
				case i := <-in:
					if isPrime(i.(int)) {
						// Add some delay to make the time noticeable.
						time.Sleep(500 * time.Millisecond)
						select {
						case <-done:
							return
						case outStream <- i:
						}
					}
				}
			}
		}()
		return outStream
	}

	repeatFn := func(
		done chan interface{},
		fn func() interface{},
	) <-chan interface{} {
		outStream := make(chan interface{})
		go func() {
			defer close(outStream)
			for {
				select {
				case <-done:
					return
				case outStream <- fn():
				}
			}
		}()
		return outStream

	}

	take := func(done chan interface{}, in <-chan interface{}, n int) <-chan interface{} {
		outStream := make(chan interface{})
		go func() {
			defer close(outStream)
			for i := 0; i < n; i++ {
				select {
				case <-done:
					return
				case outStream <- <-in:
				}
			}
		}()
		return outStream
	}

	done := make(chan interface{})
	defer close(done)

	result := take(done, primeFinder(done, repeatFn(done, randFn)), 10)
	before := time.Now()
	for res := range result {
		fmt.Println(res)
	}
	fmt.Println("took", time.Since(before))

	fanOut := func(
		primeFinder PrimeFinder,
		done chan interface{},
		in <-chan interface{},
	) []<-chan interface{} {
		n := runtime.NumCPU()
		n = 10
		workers := make([]<-chan interface{}, n)
		for i := 0; i < n; i++ {
			workers[i] = primeFinder(done, in)
		}
		return workers
	}

	fanIn := func(
		done chan interface{},
		channels ...<-chan interface{},
	) chan interface{} {
		var wg sync.WaitGroup
		outStream := make(chan interface{})

		multiplex := func(in <-chan interface{}) {
			defer wg.Done()
			for i := range in {
				select {
				case <-done:
					return
				case outStream <- i:
				}
			}

		}
		wg.Add(len(channels))
		for _, channel := range channels {
			go multiplex(channel)

		}

		go func() {
			wg.Wait()
			close(outStream)
		}()

		return outStream
	}

	before = time.Now()
	channels := fanOut(primeFinder, done, repeatFn(done, randFn))
	result = take(done, fanIn(done, channels...), 10)
	for res := range result {
		fmt.Println(res)
	}
	fmt.Println("took", time.Since(before))
}
```
