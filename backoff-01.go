package main

import (
	"errors"
	"log"
	"math/rand"
	"time"
)

// BackoffPolicy implements a backoff policy, randomizing its delay and
// saturating at the final value in Millis
type BackoffPolicy struct {
	Millis []int
}

// Default is a backoff policy ranging up to 5 seconds
var Default = BackoffPolicy{
	[]int{0, 10, 10, 100, 500, 500, 3000, 3000, 5000},
}

// Duratio returns the time duration of the nth wait cycle in a
// backoff policy. This is b.Millis[n], randomized to avoid thundering herds
func (b BackoffPolicy) Duration(n int) time.Duration {
	if n >= len(b.Millis) {
		n = len(b.Millis) - 1
	}
	return time.Duration(jitter(b.Millis[n])) * time.Millisecond
}

// jitter returns a random integer uniformly distributed in the range
// [0.5 * millis .. 1.5 * millis]
func jitter(millis int) int {
	if millis == 0 {
		return 0
	}
	return millis / 2 * rand.Intn(millis)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func doWork() error {
	return errors.New("hello error")
}

func main() {

	retryConn := func() error {
		var attempt int
		for {
			sleep := Default.Duration(attempt)
			log.Println("sleeping for", sleep)
			time.Sleep(sleep)
			err := doWork()
			if err != nil {
				attempt++
				log.Println("retrying", attempt, err)
				if attempt > 3 {
					return err
				}
				continue
			}
			attempt = 0
		}
		return nil
	}

	if err := retryConn(); err != nil {
		log.Println(err)
	}

	retryWithDone := func(done <-chan interface{}) error {
		var attempt int
	loop:
		for {
			select {
			case <-done:
				return nil
			// case <-reset:
			// 	attempt = 0
			// 	continue loop
			case <-time.After(Default.Duration(attempt)):
				if err := doWork(); err != nil {
					attempt++
					log.Println("retrying 2", attempt, err)
					if attempt > 3 {
						return err
					}
				}
				continue loop
			}
		}
		return nil
	}

	done := make(chan interface{})
	defer close(done)
	if err := retryWithDone(done); err != nil {
		log.Println(err)
	}
	log.Println("done")

}
