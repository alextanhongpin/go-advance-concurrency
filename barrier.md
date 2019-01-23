A set of concurrent work must complete first, before the next batch of concurrent work can happen.
```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var mu sync.Mutex
	cond := sync.NewCond(new(sync.Mutex))

	var i int
	condition1 := func() bool {
		return i == 3
	}

	increment := func() {
		fmt.Println("increment")
		time.Sleep(1 * time.Second)
		mu.Lock()
		defer mu.Unlock()
		i++
		if i == 3 {
			cond.Broadcast()
		}
	}

	condition2 := func() bool {
		return i == 0
	}

	decrement := func() {
		fmt.Println("decrement")
		time.Sleep(1 * time.Second)
		mu.Lock()
		defer mu.Unlock()
		i--
		if i == 0 {
			cond.Broadcast()
		}
	}

	work := func(c *sync.Cond, wg *sync.WaitGroup) {

		fmt.Println("start")
		cond.L.Lock()
		go increment()
		for !condition1() {
			fmt.Println("start work 1")
			cond.Wait()
			fmt.Println("do work")
		}
		cond.L.Unlock()
		wg.Done()

		wg.Add(1)
		cond.L.Lock()
		go decrement()
		for !condition2() {
			fmt.Println("start work 2")
			cond.Wait()
			fmt.Println("do work 2")
		}
		cond.L.Unlock()
		fmt.Println("end")
		wg.Done()

	}

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go work(cond, &wg)
	}
	wg.Wait()
	fmt.Println("exiting")
}
```


## References
- https://medium.com/golangspec/reusable-barriers-in-golang-156db1f75d0b
