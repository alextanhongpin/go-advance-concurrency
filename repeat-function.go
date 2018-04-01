package main

import (
	"fmt"
	"math/rand"
)

func main() {
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

	done := make(chan interface{})
	defer close(done)

	rand := func() interface{} { return rand.Intn(10) }
	for num := range take(done, repeatFn(done, rand), 10) {
		fmt.Printf("%v ", num)
	}
}
