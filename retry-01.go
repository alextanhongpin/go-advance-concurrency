package main

import (
	"errors"
	"log"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {

	retry := func(maxRetry int, maxDuration, sleep time.Duration, callback func() error) error {
		t0 := time.Now()
		i := 0
		for {
			i++
			err := callback()
			if err == nil {
				return nil
			}
			if i >= maxRetry {
				log.Println("attempts", i)
				return errors.New("exceeded max attemps")
			}
			if time.Since(t0) > time.Duration(maxDuration)*time.Millisecond {
				log.Println(time.Since(t0), maxDuration, time.Since(t0) > maxDuration)
				return errors.New("exceeded retry time")
			}
			time.Sleep(time.Duration(sleep) * time.Millisecond)
			log.Println("retrying due to error:", err)
		}
		return nil
	}

	err := retry(10, 75, 10, func() error {
		if rand.Float32() > 0.2 {
			return errors.New("something happened")
		}
		return nil
		// return errors.New("something happened")
	})

	if err != nil {
		log.Println("irrecoverable error", err)
	}
	log.Println("done")
}
