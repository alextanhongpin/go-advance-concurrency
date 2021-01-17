## Mistake 1: Sending message to another channel


```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	ch1 := make(chan interface{})
	ch2 := make(chan interface{})
	done := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				close(ch1)
				close(ch2)
				return
			case v := <-ch1:
				fmt.Println("received from ch1", v)
				ch2 <- v
			case v := <-ch2:
				fmt.Println("received from ch2", v)
			}
		}
	}()

	go func() {
		for i := 0; i < 10; i++ {
			ch1 <- i
		}
	}()

	select {
	case <-time.After(3 * time.Second):
		close(done)
	}

	wg.Wait()
	fmt.Println("Hello, playground")
}
```

This will block at the `ch1`. To prevent blocking, either use a buffered channel for `ch2`, or separate the processing. Example of separating the processing:

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	ch1 := make(chan interface{})
	ch2 := make(chan interface{})
	done := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				close(ch1)
				return
			case v := <-ch1:
				fmt.Println("received from ch1", v)
				ch2 <- v
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				close(ch2)
				return
			case v := <-ch2:
				fmt.Println("received from ch2", v)
			}
		}
	}()

	go func() {
		for i := 0; i < 3; i++ {
			ch1 <- i
		}
	}()

	select {
	case <-time.After(3 * time.Second):
		close(done)
	}

	wg.Wait()
	fmt.Println("Hello, playground")
}
```
