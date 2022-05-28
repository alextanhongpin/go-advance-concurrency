# Promises with golang

This is a proof of concept on how to implement promises with golang. See also [promise](https://github.com/alextanhongpin/promise).

Don't use this in production.

An alternative design is to use [dataloader2](https://github.com/alextanhongpin/dataloader2).


```go
// You can edit this code!
// Click here and start typing.
package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

type Status int32

const (
	Pending Status = iota
	Rejected
	Fulfilled
)

func (s Status) Int32() int32 {
	return int32(s)
}

func (s Status) String() string {
	switch s {
	case Pending:
		return "pending"
	case Rejected:
		return "rejected"
	case Fulfilled:
		return "fulfilled"
	default:
		return "unknown"
	}
}

func (s Status) Valid() bool {
	return s >= Pending && s <= Fulfilled
}

var ErrNilPromise = errors.New("promise not initialized")

type Promise[T any] struct {
	wg     sync.WaitGroup
	res    T
	err    error
	once   sync.Once
	status Status
	cancel func()
}

type PromiseFunc[T any] func(ctx context.Context) (T, error)

func NewPromise[T any](fn PromiseFunc[T]) *Promise[T] {
	ctx, cancel := context.WithCancel(context.Background())

	p := new(Promise[T])
	p.cancel = cancel
	p.runAsync(ctx, fn)

	return p
}

func (p *Promise[T]) Await() (T, error) {
	p.wg.Wait()
	return p.Result(), p.Error()
}

func (p *Promise[T]) Abort() {
	p.cancel()
}

func (p *Promise[T]) Status() Status {
	return Status(p.status)
}

func (p *Promise[T]) IsZero() bool {
	return p == nil && p.Status() == Pending
}

func (p *Promise[T]) Result() (t T) {
	if p.IsZero() {
		return
	}
	return p.res
}

func (p *Promise[T]) Error() error {
	if p.IsZero() {
		return ErrNilPromise
	}
	return p.err
}

func (p *Promise[T]) Then(fn func(context.Context, T) (T, error)) *Promise[T] {
	return NewPromise[T](func(ctx context.Context) (T, error) {
		res, err := p.Await()
		if err != nil {
			return res, err
		}
		return fn(ctx, res)
	})
}

func (p *Promise[T]) runAsync(ctx context.Context, fn PromiseFunc[T]) {
	p.wg.Add(1)

	go func() {
		defer p.wg.Done()
		defer p.Abort()

		res, err := fn(ctx)
		if err != nil {
			p.reject(err)
		} else {
			p.resolve(res)
		}
	}()
}

func (p *Promise[T]) resolve(res T) {
	p.once.Do(func() {
		p.status = Fulfilled
		p.res = res
	})
}

func (p *Promise[T]) reject(err error) {
	p.once.Do(func() {
		p.status = Rejected
		p.err = err
	})
}

func All[T any](promises ...*Promise[T]) *Promise[[]T] {
	if len(promises) == 0 {
		return nil
	}

	return NewPromise(func(ctx context.Context) ([]T, error) {
		n := int64(runtime.NumCPU())
		sem := semaphore.NewWeighted(n)

		result := make([]T, len(promises))

		var once sync.Once
		var firstErr error
		var wg sync.WaitGroup

		for i, p := range promises {
			if firstErr != nil {
				break
			}

			wg.Add(1)
			sem.Acquire(context.Background(), 1)

			go func(i int, p *Promise[T]) {
				defer sem.Release(1)
				defer wg.Done()

				res, err := p.Await()
				if err != nil {
					once.Do(func() {
						firstErr = err
						for j := range promises {
							promises[j].Abort()
						}
					})
				} else {
					result[i] = res
				}
			}(i, p)
		}

		wg.Wait()

		if firstErr != nil {
			return nil, firstErr
		}
		return result, nil
	})
}

func main() {
	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()

	p := NewPromise(func(ctx context.Context) (int, error) {
		fmt.Println("work 1")
		select {
		case <-time.After(1 * time.Second):
			return 42, nil
		case <-ctx.Done():
			fmt.Println("aborting")
			return -1, ctx.Err()
		}
	})
	go func() {
		time.Sleep(1500 * time.Millisecond)
		p.Abort()
	}()
	p2 := p.Then(func(ctx context.Context, n int) (int, error) {
		fmt.Println("work 2")
		select {
		case <-time.After(1 * time.Second):
			return n + 100, nil
		case <-ctx.Done():
			return -1, ctx.Err()
		}
	})
	go func() {
		time.Sleep(1500 * time.Millisecond)
		p.Abort()
	}()
	res, err := p2.Await()
	fmt.Println(res, err)

	results, err := All[int](
		NewPromise(fetchNumber),
		NewPromise(fetchNumber),
		NewPromise(fetchNumber),
		NewPromise(fetchNumber),
		NewPromise(fetchNumber),
	).Await()
	fmt.Println(results, err)
}

func fetchNumber(ctx context.Context) (int, error) {
	n := rand.Intn(10)
	if rand.Intn(2) == 1 {
		fmt.Println("bad worker", n)
		return 0, errors.New("bad luck")
	}
	fmt.Println("good worker", n)
	select {
	case <-time.After(3 * time.Second):
		fmt.Println("good worker done", n)
		return n, nil
	case <-ctx.Done():
		fmt.Println("good worker aborted", n)
		return -1, ctx.Err()
	}
}
```
