package main

import "log"

func main() {
	repeat := func(
		done <-chan interface{},
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
		done <-chan interface{},
		valueStream <-chan interface{},
		num int,
	) <-chan interface{} {
		takeStream := make(chan interface{})

		go func() {
			defer close(takeStream)

			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case takeStream <- <-valueStream:
				}
			}
		}()

		return takeStream
	}

	toString := func(
		done <-chan interface{},
		valueStream <-chan interface{},
	) <-chan string {

		stringStream := make(chan string)

		go func() {
			defer close(stringStream)

			for v := range valueStream {
				select {
				case <-done:
					return
				case stringStream <- v.(string):
				}
			}
		}()

		return stringStream
	}

	done := make(chan interface{})
	defer close(done)

	var message string
	for msg := range toString(done, take(done, repeat(done, "I", "must"), 10)) {
		message += " " + msg

	}

	log.Println(message)
}
