package main

import (
	"errors"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Job struct {
	run   func() error
	retry int
	index int
}

func work() error {
	if rand.Float32() > 0.2 {
		return errors.New("an error occured")
	}
	return nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {

	concurrent := 3

	producer := func(done <-chan interface{}) chan Job {
		jobs := make(chan Job, 10)
		go func() {
			for i := 0; i < 10; i++ {
				job := Job{
					run:   work,
					retry: 3,
					index: i,
				}
				select {
				case <-done:
					return
				case jobs <- job:
				}
			}
		}()
		return jobs
	}

	worker := func(i int, done <-chan interface{}, count chan interface{}, jobs chan Job) {

		go func() {
			// defer close(jobs)
			// defer close(count)
			for {
				select {
				case <-done:
					return
				case j, ok := <-jobs:
					if !ok {
						log.Println("closing job channels")
						return
					}
					if err := j.run(); err != nil {
						j.retry--
						if j.retry > 0 {
							log.Printf("retrying %d, attempt %d worker %d", j.index, 3-j.retry, i)
							// Push back to the jobs channel
							jobs <- j
						} else {
							count <- j.index
						}
					} else {
						count <- j.index
					}
				}
			}
		}()

	}

	cleaner := func(done <-chan interface{}, count chan interface{}, jobs chan Job) {
		var wg sync.WaitGroup
		wg.Add(10)
		go func() {
			for i := 0; i < 10; i++ {
				select {
				case <-done:
					return
				default:
					log.Println("got", <-count)
					wg.Done()
				}
			}
			close(count)
			close(jobs)
		}()
		wg.Wait()
		log.Println("Complete")
	}

	done := make(chan interface{})
	defer close(done)

	prod := producer(done)
	count := make(chan interface{}, 10)
	for i := 0; i < concurrent; i++ {
		worker(i, done, count, prod)
	}
	cleaner(done, count, prod)
	log.Println("done")
}
