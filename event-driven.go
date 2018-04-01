package main

import (
	"log"
	"math/rand"
	"sync"
)

func main() {

	generator := func(done <-chan interface{}, upperLimit int) <-chan interface{} {
		valueStream := make(chan interface{})

		go func() {
			defer close(valueStream)

			for {
				select {
				case <-done:
					return
				case valueStream <- rand.Intn(upperLimit):
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

	orDone := func(
		done <-chan interface{},
		valueStream <-chan interface{},
	) <-chan interface{} {
		doneStream := make(chan interface{})

		go func() {
			defer close(doneStream)
			for {
				select {
				case <-done:
					return
				case val, ok := <-valueStream:
					if ok == false {
						return
					}
					select {
					case doneStream <- val:
					case <-done:
					}
				}
			}
		}()

		return doneStream
	}

	copier := func(
		done <-chan interface{},
		valueStream <-chan interface{},
	) (_, _ <-chan interface{}) {
		out1 := make(chan interface{})
		out2 := make(chan interface{})

		go func() {
			defer close(out1)
			defer close(out2)

			for val := range orDone(done, valueStream) {
				var out1, out2 = out1, out2
				for i := 0; i < 2; i++ {
					select {
					case <-done:
						return
					case out1 <- val:
						out1 = nil
					case out2 <- val:
						out2 = nil
					}
				}
			}
		}()

		return out1, out2
	}

	filter := func(
		done <-chan interface{},
		valueStream <-chan interface{},
		condition func(interface{}) bool,
	) <-chan interface{} {
		filterStream := make(chan interface{})

		go func() {
			defer close(filterStream)

			for val := range orDone(done, valueStream) {
				if condition(val) {
					select {
					case <-done:
					case filterStream <- val:
					}
				}
			}
		}()

		return filterStream
	}

	merger := func(
		done <-chan interface{},
		channels ...<-chan interface{},
	) <-chan interface{} {
		mergeStream := make(chan interface{})

		var wg sync.WaitGroup

		multiplex := func(c <-chan interface{}) {
			defer wg.Done()
			for i := range c {
				select {
				case <-done:
					return
				case mergeStream <- i:
				}
			}
		}

		wg.Add(len(channels))

		for _, c := range channels {
			go multiplex(c)
		}

		go func() {
			defer close(mergeStream)
			wg.Wait()
		}()

		return mergeStream
	}

	toList := func(
		done <-chan interface{},
		valueStream <-chan interface{},
	) []interface{} {
		var list []interface{}

		for v := range valueStream {
			select {
			case <-done:
			default:
				list = append(list, v)
			}
		}

		return list
	}

	// FILTERS
	isEven := func(val interface{}) bool {
		return val.(int)%2 == 0
	}

	isLess50 := func(val interface{}) bool {
		return val.(int) < 50
	}

	isGreater500 := func(val interface{}) bool {
		return val.(int) > 500
	}

	// START

	done := make(chan interface{})
	defer close(done)

	genPipe := generator(done, 1000)
	filterPipe := filter(done, genPipe, isEven)
	takePipe := take(done, filterPipe, 50)
	a, b := copier(done, takePipe)

	c := merger(done, filter(done, a, isLess50), filter(done, b, isGreater500))

	output := toList(done, c)
	log.Println(output)

}
