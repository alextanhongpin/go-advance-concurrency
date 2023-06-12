```go
// You can edit this code!
// Click here and start typing.
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func main() {
	ctx := context.Background()
	cancel := lock(ctx)
	time.Sleep(1 * time.Second)
	fmt.Println(cancel())

	fmt.Println("Hello, 世界")
}

func lock(ctx context.Context) func() error {
	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan error)
	go func() {

		fmt.Println("start")
		<-ctx.Done()
		ch <- ctx.Err()
	}()

	return func() error {
		cancel()
		err := <-ch
		fmt.Println("end")
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
}
```
