# Repeat, and Repeat Pipeline
```go
package main

import (
	"fmt"
	"math/rand"
)

func main() {

	repeat := func(
		done chan interface{},
		values ...interface{},
	) <-chan interface{} {
		valueStream := make(chan interface{})
		go func() {
			defer close(valueStream)
			for {
				for _, v := range values {
					select {
					case <-done:
						return
					case valueStream <- v:
					}

				}
			}
		}()
		return valueStream
	}

	take := func(
		done chan interface{},
		in <-chan interface{},
		n int,
	) <-chan interface{} {
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

	done := make(chan interface{})
	alphabets := []interface{}{21, 42}
	result := take(done, repeat(done, alphabets...), 10)
	for res := range result {
		fmt.Println(res)
	}

	randFn := func() interface{} {
		return rand.Int()
	}
	result = take(done, repeatFn(done, randFn), 10)
	for res := range result {
		fmt.Println(res)
	}
}
```
