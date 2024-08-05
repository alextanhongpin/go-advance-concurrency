# Refresh pattern

Runs two separate goroutine
- one for performing the task
- another for refreshing, e.g. can be a heartbeat, ping, or even extending a lock periodically

If the task is completed, cancel the refresh operation without error.
Otherwise, if the task or refresh failed, the whole operation should be cancelled.


```go
// You can edit this code!
// Click here and start typing.
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

var ErrCompleted = errors.New("completed")

func main() {
	fmt.Println("got", work(context.Background()))
}

func work(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancelCause(ctx)

	g.Go(func() error {
		err := refresh(ctx)
		if errors.Is(err, ErrCompleted) {
			return nil
		}

		return err
	})
	g.Go(func() error {
		defer cancel(ErrCompleted)

		return perform(ctx)
	})

	return g.Wait()
}

// refresh can be a heartbeat, ping, or extending a lock etc.
func refresh(ctx context.Context) error {
	for i := range 10 {
		select {
		case <-ctx.Done():
			fmt.Println("refresh: done", context.Cause(ctx))
			return fmt.Errorf("refresh: %w", context.Cause(ctx))
		default:
			fmt.Println("refresh: do")
			if i > 3 {
				fmt.Println("refresh: failed")
				return errors.New("refresh: error")
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil
}

func perform(ctx context.Context) error {
	select {
	case <-ctx.Done():
		fmt.Println("perform: done", context.Cause(ctx))
		return fmt.Errorf("perform: %w", context.Cause(ctx))
	case <-time.After(550 * time.Millisecond):
		fmt.Println("perform: do")
		if false {
			return errors.New("perform: error")
		}
		return nil
	}
}
```
