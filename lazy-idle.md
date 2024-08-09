
In this example, we demonstrate lazy initialization of goroutine, as well as idle termination of goroutine.

- lazy initialization - instead of starting the goroutine immediately during construction, we delay it to start only when the first request that requires the goroutine is executed. This is useful if you do not want to spawn too many goroutines that will be left unused.
- idle termination - similarly, if the goroutine is idle after the idle timeout, we terminate it. It can be started again by sending another request. If a new request enters before the idle timeout, the timeout will be extended.


```go
package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	b := New(func(v int) {
		time.Sleep(10 * time.Millisecond)
		fmt.Println("recv:", v)
	}, &Options{
		IdleTimeout: 500 * time.Millisecond,
	})
	defer b.Stop()

	b.Send(1)
	b.Send(2)
	time.Sleep(300 * time.Millisecond)
	b.Send(3)
	time.Sleep(time.Second)
	b.Send(4)
	b.Send(5)
}

type Background[T any] struct {
	ch          chan T
	idleTimeout time.Duration
	awake       atomic.Bool // Whether the goroutine has been started.
	done        chan struct{}
	wg          sync.WaitGroup
	fn          func(T)
	end         sync.Once
}

type Options struct {
	IdleTimeout time.Duration
}

func New[T any](fn func(T), opts *Options) *Background[T] {
	return &Background[T]{
		ch:          make(chan T),
		idleTimeout: opts.IdleTimeout,
		done:        make(chan struct{}),
		fn:          fn,
	}
}

func (b *Background[T]) Stop() {
	b.end.Do(func() {
		close(b.done)

		b.wg.Wait()
	})
}

func (b *Background[T]) Send(v T) bool {
	// We only start on demand.
	b.start()

	select {
	case <-b.done:
		return false
	case b.ch <- v:
		return true
	}
}

func (b *Background[T]) wake() bool {
	prev := b.awake.Swap(true)
	return !prev
}

func (b *Background[T]) sleep() {
	fmt.Println("sleeping...")
	b.awake.Swap(false)
}

func (b *Background[T]) start() {
	select {
	case <-b.done:
		return
	default:
	}

	// Wake the goroutine if it's not running.
	// If it is already awake, don't start a new goroutine.
	if ok := b.wake(); !ok {
		fmt.Println("already awake")
		return
	}

	fmt.Println("waking up goroutine...")

	// Start the goroutine.
	b.wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer b.wg.Done()
		defer cancel()
		defer b.sleep()

		b.loop(ctx)
	}()
}

func (b *Background[T]) loop(ctx context.Context) {
	idle := time.NewTicker(b.idleTimeout)
	defer idle.Stop()

	for {
		select {
		case <-b.done:
			return
		case <-idle.C:
			return
		case v := <-b.ch:
			// Reset the idle timeout.
			idle.Reset(b.idleTimeout)

			b.fn(v)
		}
	}
}
```
