```go
// You can edit this code!
// Click here and start typing.
package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

func main() {
	doer := backpressure(1, 1, time.Second)
	n := 5
	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			fmt.Println(doer(func() error {
				time.Sleep(time.Second)
				return nil
			}))
		}()
	}
	wg.Wait()
	fmt.Println("Hello, 世界")
}

func backpressure(limit, backlogLimit int, backlogTimeout time.Duration) func(fn func() error) error {
	ch := make(chan struct{}, limit)
	bch := make(chan struct{}, limit+backlogLimit)
	for range limit {
		ch <- struct{}{}
	}
	for range limit + backlogLimit {
		bch <- struct{}{}
	}

	return func(fn func() error) error {
		select {
		case <-bch:
			defer func() {
				bch <- struct{}{}
			}()
			select {
			case <-time.After(backlogTimeout):
				return errors.New("timeout")
			case <-ch:
				defer func() {
					ch <- struct{}{}
				}()
				return fn()
			default:
				return errors.New("cap exceeded")
			}
		default:
			return errors.New("backlog exceeded")
		}
	}
}

```
