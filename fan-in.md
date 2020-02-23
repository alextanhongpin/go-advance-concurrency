## Fan-in and Fan-out

From the [Go Concurrency Patterns: Pipelines](https://blog.golang.org/pipelines):

_Fan-in_: Multiple functions can read from the same channel until that channel is closed. This provides a way to distribute work amongst a group of workers to parallelize CPU use and I/O.
_Fan-out_: A function can read from multiple inputs and proceed until all are closed by multiplexing the input channels onto a single channel that's closed when all the inputs are closed.

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

func isPrime(n int) bool {
	// 0 and 1 is not a prime number.
	if n < 2 {
		return false
	}

	// 2 is a prime number.
	if n == 2 {
		return true
	}
	// Any number divisible by 2 is not a prime number.
	if n > 2 && n%2 == 0 {
		return false
	}

	maxDiv := int(math.Floor(math.Sqrt(float64(n))))
	for i := 3; i < maxDiv+1; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func repeatFn(
	done chan interface{},
	fn func() interface{},
) chan interface{} {
	stream := make(chan interface{})

	go func() {
		defer close(stream)

		for {
			select {
			case <-done:
				return
			case stream <- fn():
			}
		}
	}()
	return stream
}

func takeInt(
	done chan interface{},
	valueStream chan int,
	n int,
) chan int {
	stream := make(chan int)
	go func() {
		defer close(stream)

		for i := 0; i < n; i++ {
			select {
			case <-done:
				return
			case stream <- <-valueStream:
			}
		}
	}()
	return stream
}

func toInt(done chan interface{}, valueStream chan interface{}) chan int {
	intStream := make(chan int)

	go func() {
		defer close(intStream)

		for {
			select {
			case <-done:
				return
			case n := <-valueStream:
				v, ok := n.(int)
				if ok {
					intStream <- v
				}
			}
		}
	}()

	return intStream
}

func primeFinder(done chan interface{}, intStream chan int) chan int {
	outStream := make(chan int)

	go func() {
		defer close(outStream)
		for {
			select {
			case <-done:
				return
			case n := <-intStream:
				if isPrime(n) {
					time.Sleep(100 * time.Millisecond)
					outStream <- n
				}
			}
		}
	}()
	return outStream
}

func fanIn(
	done chan interface{},
	channels ...chan int,
) chan int {
	stream := make(chan int)

	var wg sync.WaitGroup

	multiplex := func(ch chan int) {
		defer wg.Done()
		for v := range ch {
			select {
			case <-done:
				return
			case stream <- v:

			}
		}
	}

	wg.Add(len(channels))
	for i := 0; i < len(channels); i++ {
		go multiplex(channels[i])
	}

	go func() {
		wg.Wait()
		close(stream)
	}()

	return stream
}

func main() {
	randFn := func() interface{} { return rand.Intn(50_000) }

	done := make(chan interface{})
	defer close(done)
	var concurrent bool
	// Toggle this to see the difference.
	concurrent = true
	if concurrent {
		n := runtime.NumCPU()
		channels := make([]chan int, n)
		for i := 0; i < n; i++ {
			channels[i] = primeFinder(done, toInt(done, repeatFn(done, randFn)))
		}

		start := time.Now()
		for v := range takeInt(done, fanIn(done, channels...), 10) {
			fmt.Println(v)
		}
		fmt.Printf("running with %d workers: %s", n, time.Since(start))
	} else {
		fn := primeFinder(done, toInt(done, repeatFn(done, randFn)))

		start := time.Now()
		for v := range takeInt(done, fn, 10) {
			fmt.Println(v)
		}
		fmt.Printf("running single worker: %s", time.Since(start))
	}
}
```
