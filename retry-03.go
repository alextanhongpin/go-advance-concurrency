package main

import (
	"errors"
	"log"
	"math/rand"
	"time"
)

type Retry struct {
	MaxDuration time.Duration
	MaxRetry    int
	Count       int
	SleepTime   int
}

func (r *Retry) Sleep() {
	r.Count++
	jitter := time.Duration(rand.Intn(r.SleepTime)) * time.Millisecond
	sleep := time.Duration(r.SleepTime) + jitter/2

	log.Println("jitter", sleep, time.Duration(sleep*2))
	time.Sleep(sleep * time.Duration(r.Count))
}

func (r *Retry) ThresholdExceeded() bool {
	return r.Count >= r.MaxRetry
}

func NewRetry(maxRetry, maxDuration int) *Retry {
	return &Retry{
		MaxRetry:    maxRetry,
		MaxDuration: time.Duration(maxDuration) * time.Millisecond,
		Count:       0,
		SleepTime:   500,
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func doWork() error {
	log.Println("doing work")
	return errors.New("unable to perform work")
}

func doWorkWithRetry() error {

	retry := NewRetry(3, 100)
	for {
		if err := doWork(); err != nil {
			retry.Sleep()
			if retry.ThresholdExceeded() {
				return err
			}
		}
	}
	return nil
}

func main() {
	if err := doWorkWithRetry(); err != nil {
		log.Println(err)
	}
}
