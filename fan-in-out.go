package main

import (
	"log"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

func main() {

	isPrime := func(value int) bool {
		for i := 2; i <= int(math.Floor(math.Sqrt(float64(value)))); i++ {
			if value%i == 0 {
				return false
			}
		}
		return value > 1
	}

	primeFinder := func(
		done <-chan interface{},
		valueStream <-chan int,
	) <-chan interface{} {
		primeStream := make(chan interface{})

		go func() {
			defer close(primeStream)

			for {
				v := <-valueStream
				if isPrime(v) {
					select {
					case <-done:
						return
					case primeStream <- v:
					}
				}
			}
		}()
		return primeStream
	}

	repeatFn := func(
		done <-chan interface{},
		fn func() interface{},
	) <-chan interface{} {
		valueStream := make(chan interface{})

		go func() {
			defer close(valueStream)
			for {
				select {
				case <-done:
					return
				case valueStream <- fn():
				}
			}
		}()

		return valueStream
	}

	take := func(
		done <-chan interface{},
		valueStream <-chan interface{},
		limit int,
	) <-chan interface{} {
		takeStream := make(chan interface{})

		go func() {
			defer close(takeStream)
			for i := 0; i < limit; i++ {
				select {
				case <-done:
					return
				case takeStream <- <-valueStream:
				}
			}
		}()

		return takeStream
	}

	toInt := func(
		done <-chan interface{},
		valueStream <-chan interface{},
	) <-chan int {
		intStream := make(chan int)

		go func() {
			defer close(intStream)

			for {
				v := <-valueStream
				select {
				case <-done:
					return
				case intStream <- v.(int):
				}
			}
		}()

		return intStream
	}

	fanIn := func(
		done <-chan interface{},
		channels ...<-chan interface{},
	) <-chan interface{} {
		var wg sync.WaitGroup
		wg.Add(len(channels))

		multiplexedStream := make(chan interface{})

		multiplex := func(c <-chan interface{}) {
			defer wg.Done()
			for i := range c {
				select {
				case <-done:
					return
				case multiplexedStream <- i:
				}
			}
		}

		for _, c := range channels {
			go multiplex(c)
		}

		go func() {
			defer close(multiplexedStream)
			wg.Wait()
		}()

		return multiplexedStream
	}

	done := make(chan interface{})
	defer close(done)

	rand := func() interface{} { return rand.Intn(500000000) }
	// NOTE: Without fan-in/fan-out: 3.301513ms
	// With fan-in/fan-out: time since: 1.25219ms

	start := time.Now()
	randIntStream := toInt(done, repeatFn(done, rand))

	// for i := range take(done, primeFinder(done, randIntStream), 10) {
	// 	log.Println(i)
	// }
	// log.Printf("time since: %v\n", time.Since(start))
	numFinders := runtime.NumCPU()
	finders := make([]<-chan interface{}, numFinders)
	for i := 0; i < numFinders; i++ {
		finders[i] = primeFinder(done, randIntStream)
	}

	for prime := range take(done, fanIn(done, finders...), 10) {
		log.Println(prime)
	}
	log.Printf("time since: %v\n", time.Since(start))
}
