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
