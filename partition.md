# How to partition data cross channels
```go
package main

import (
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

func main() {
	n := 4

	var wg sync.WaitGroup
	wg.Add(n)

	channels := make([]chan int, n)
	for i := 0; i < n; i++ {
		channels[i] = make(chan int)
	}

	for i, ch := range channels {
		go process(&wg, i, ch)
	}

	h := fnv.New64()
	for i := 0; i < 10; i++ {
		h.Write([]byte(fmt.Sprint(i)))

		channels[int(h.Sum64()%uint64(n))] <- i
		h.Reset()
	}

	time.Sleep(1 * time.Second)
	for i := range channels {
		close(channels[i])
	}

	wg.Wait()
}

func process(wg *sync.WaitGroup, i int, ch chan int) {
	defer wg.Done()

	for n := range ch {
		fmt.Printf("#worker%d: %d\n", i, n)
	}
}
```
