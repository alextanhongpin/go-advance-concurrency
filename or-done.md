# Basic 
```go
package main

import (
	"fmt"
)

func main() {
	orDone := func(done, c <-chan interface{}) <-chan interface{} {
		valStream := make(chan interface{})

		go func() {
			defer close(valStream)
			for {
				select {
				case <-done:
					return
				case v, ok := <-c:
					if ok == false {
						return
					}
					select {
					case valStream <- v:
					case <-done:
					}
				}
			}
		}()
		return valStream
	}

	generators := func(
		done <-chan interface{},
		integers ...int,
	) <-chan interface{} {
		valueStream := make(chan interface{})

		go func() {
			defer close(valueStream)
			for _, i := range integers {
				select {
				case <-done:
					return
				case valueStream <- i:
				}
			}
		}()

		return valueStream
	}

	done := make(chan interface{})
	for val := range orDone(done, generators(done, 1, 2, 3, 4, 5)) {
		fmt.Print(val)
	}
}
```

## With closed channel

```go
package main

import "fmt"

func orDone(
	done chan interface{},
	in <-chan interface{},
) <-chan interface{} {
	outStream := make(chan interface{})
	go func() {
		defer close(outStream)
		for {
			select {
			case <-done:
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case outStream <- v:
				case <-done:
				}
			}
		}
	}()
	return outStream
}

func main() {
	done := make(chan interface{})
	defer close(done)

	in := make(chan interface{})
	close(in)
	
	result := orDone(done, in)
	for res := range result {
		fmt.Println(res)
	}
}
```
