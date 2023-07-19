# Circuit Breaker using Channel

## Status

`draft`

## Context

We want to implement circuit breaker capability while applying golang concurrency.

## Decision

We will use channels instead of mutexes.

## Consequences

A working example of circuit breaker without locking.


```go
package main

import (
	 "errors"
	 "fmt"
	 "sync"
	 "time"
)

func main() {
	 cb := New(5, 1000*time.Millisecond)
	 defer cb.Close()

	 for i := 0; i < 10; i++ {
			 err := cb.Do(func() error {
					 return errors.New("boo")
			 })
			 fmt.Println(err)
	 }
	 time.Sleep(1 * time.Second)
	 for i := 0; i < 10; i++ {
			 err := cb.Do(func() error {
					 return nil
			 })
			 fmt.Println(err)
	 }
	 fmt.Println("Hello, 世界")
}

type CircuitBreaker struct {
	 done      chan any
	 count     int
	 ch        chan error
	 threshold int
	 sleep     time.Duration
	 wg        sync.WaitGroup
}

func New(threshold int, sleep time.Duration) *CircuitBreaker {
	 cb := &CircuitBreaker{
			 threshold: threshold,
			 ch:        make(chan error),
			 sleep:     sleep,
			 done:      make(chan any),
	 }

	 cb.init()
	 return cb
}

func (c *CircuitBreaker) Close() {
	 close(c.ch)
	 c.wg.Wait()
}

func (c *CircuitBreaker) init() {
	 c.wg.Add(1)
	 go func() {
			 defer c.wg.Done()
			 for {
					 select {
					 case err, ok := <-c.ch:
							 if !ok {
									 return
							 }

							 select {
							 case <-c.done:
									 continue
							 default:
									 if err == nil {
											 if c.count > 0 {
													 c.count--
											 }
									 } else {
											 if c.count < c.threshold {
													 c.count++
											 } else if c.count == c.threshold {
													 close(c.done)
													 c.reset()
											 }
									 }
							 }
					 }
			 }
	 }()
}

func (c *CircuitBreaker) reset() {
	 time.AfterFunc(c.sleep, func() {
			 c.done = make(chan any)
	 })
}

func (c *CircuitBreaker) Do(fn func() error) error {
	 select {
	 case <-c.done:
			 return errors.New("too many requests")
	 default:
			 err := fn()
			 c.ch <- err
			 return err
	 }
}
```
