## Multiple consumer

We can have multiple consumers to a single producer channel, but the ordering is no longer guaranteed (hence concurrent processing). If we only use a single consumer, the order will almost always be guaranteed.

```go
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	ch := make(chan interface{})

	n := 3
	var wg sync.WaitGroup
	wg.Add(n)

	listen := func(i int) {
		defer wg.Done()

		for msg := range ch {
			fmt.Printf("consumer#%d: %d\n", i, msg)
		}
	}
	for i := 0; i < n; i++ {
		go func(i int) {
			listen(i)
		}(i)
	}

	for i := 0; i < 10; i++ {
		go func(i int) {
			time.Sleep(time.Duration(rand.Intn(300)+200) * time.Millisecond)
			ch <- i
		}(i)
	}
	time.Sleep(3 * time.Second)

	close(ch)
	wg.Wait()
	fmt.Println("Hello, playground")
}
```
