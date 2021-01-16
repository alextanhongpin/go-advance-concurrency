## Cancelling multiple task with context

```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {

	var wg sync.WaitGroup
	wg.Add(3)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for i := 0; i < 3; i++ {
		go func() {
			defer wg.Done()
			task(ctx)
		}()
	}
	wg.Wait()
}

func task(ctx context.Context) {
	fmt.Println("start task")
	select {
	case <-time.After(10 * time.Second):
		fmt.Println("timeout")
		return
	case <-ctx.Done():
		fmt.Println("terminating task")
		return
	}
}
```


## Use context instead of channel to signal completion

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	done := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println(ctx.Err())
				close(done)
				return
			default:
				time.Sleep(300 * time.Millisecond)
				fmt.Println("doing work")
			}
		}
	}()

	go func() {
		select {
		case <-time.After(3 * time.Second):
			cancel()
		}
	}()
	<-done
	fmt.Println("Hello, playground")
}
```
