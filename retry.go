package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(42)
}

func doWork() (int, error) {
	if n := rand.Intn(10); n == 9 {
		fmt.Println("found the lucky number")
		return n, nil
	}

	return 0, errors.New("something went wrong")
}

func main() {
	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()

	//res, err := Retry(Fixed, 10, doWork)
	//res, err := Retry(Linear, 10, doWork)
	res, err := Retry(Backoff, 10, doWork)
	if err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println("result:", res)
	log.Println("program terminating")
}

// RetryPolicy implements a backoff policy, randomizing its delay and
// saturating at the final value in Millis
type RetryPolicy struct {
	millis []int
	jitter bool
}

// Duration returns the duration at attempt n.
func (b RetryPolicy) Duration(n int) time.Duration {
	switch {
	case n < 0:
		n = 0
	case n >= len(b.millis):
		n = len(b.millis) - 1
	}

	duration := b.millis[n] / 10
	if b.jitter && duration > 0 {
		duration = duration/2 + rand.Intn(duration)
	}

	return time.Duration(duration) * time.Millisecond
}

type RetryFunc[T any] func() (T, error)

type retryPolicy interface {
	Duration(n int) time.Duration
}

var Linear = RetryPolicy{
	millis: []int{0, 1_000, 1_500, 5_000, 10_000, 25_000, 60_000},
}

var Fixed = RetryPolicy{
	millis: []int{0, 5_000, 5_000, 5_000, 5_000, 5_000, 5_000},
}

var Backoff = RetryPolicy{
	millis: []int{0, 500, 1_000, 2_500, 5_000, 15_000, 30_000},
	jitter: true,
}

func Retry[P retryPolicy, T any](p P, n int, fn RetryFunc[T]) (t T, err error) {
	for i := 0; i < n+1; i++ {
		// Remove this in production.
		fmt.Println("sleeping for", p.Duration(i), "attempt", i)

		time.Sleep(p.Duration(i))
		t, err = fn()
		if err == nil {
			return
		}
	}

	return t, fmt.Errorf("%w: too many retries", err)
}
