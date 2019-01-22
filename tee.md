# Sample Tee

Useful to duplicate result:

```go
package main

import (
	"fmt"
)

func tee(
	done <-chan interface{},
	in <-chan interface{},
) (_, _ <-chan interface{}) {
	out1 := make(chan interface{})
	out2 := make(chan interface{})

	go func() {
		defer close(out1)
		defer close(out2)
		for i := range in {
			var out1, out2 = out1, out2
			for j := 0; j < 2; j++ {
				select {
				case <-done:
					return
				case out1 <- i:
					out1 = nil
				case out2 <- i:
					out2 = nil
				}
			}
		}
	}()
	return out1, out2
}

func generator(
	done chan interface{},
	values ...interface{},
) <-chan interface{} {

	outStream := make(chan interface{})
	go func() {
		defer close(outStream)
		for _, v := range values {
			outStream <- v
		}
	}()
	return outStream
}

func main() {
	done := make(chan interface{})
	defer close(done)

	nums := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9}
	out1, out2 := tee(done, generator(done, nums...))

	for res := range out1 {
		fmt.Println(res, <-out2)
	}
}
```
