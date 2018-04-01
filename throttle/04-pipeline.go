package main

import (
	"log"
	"sync"
	"time"
)

func main() {
	numCPU := 4
	sampleSize := 12

	wg := new(sync.WaitGroup)
	in := make(chan time.Time, sampleSize)
	for i := 0; i < numCPU; i++ {
		wg.Add(1)
		go consumerWorker(in, wg)
	}

	for i := 0; i < sampleSize; i++ {
		in <- time.Now()
	}
	close(in)

	wg.Wait()
}

func consumerWorker(in <-chan time.Time, wg *sync.WaitGroup) {
	for {
		start, more := <-in
		if !more {
			break
		}

		time.Sleep(500 * time.Millisecond)
		log.Println(time.Now().Sub(start).Seconds())
	}
	wg.Done()
}
