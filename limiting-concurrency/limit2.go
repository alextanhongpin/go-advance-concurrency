package main

import (
	"log"
	"time"
)

func asyncWork() {
	time.Sleep(500 * time.Millisecond)
}

func main() {
	maxConcurrent := 5
	numberOfJobs := 20

	// Dummy channel to coordinate the number of concurrent goroutines
	// This channel should be buffered otherwise we will immediately
	// blocked when trying to fill it
	semaphore := make(chan struct{}, maxConcurrent)

	// Fill the channel with the max number of concurrent goroutines
	log.Println("Start semaphore" , len(semaphore))
	for i := 0; i < maxConcurrent; i++ {
		semaphore <- struct{}{}
	}
	log.Println("Filled semaphore" , len(semaphore))

	// This done channel indicates when a single goroutine has finished its job
	done := make(chan bool)

	// The wait all channel allow the main program to wait untilwe have done all the job
	waitAll := make(chan bool)

	// Collect all the jobs, and since the job is finished, we can release
	// another spot for a goroutine
	go func() {
		for i := 0; i < numberOfJobs; i++ {
			<-done
			log.Println("Before fill semaphore", len(semaphore), cap(semaphore))
			semaphore <- struct{}{}
			log.Println("Fill semaphore" , len(semaphore), cap(semaphore))
		}
		waitAll <- true
	}()

	// Try to start number of jobs
	for i := 1; i <= numberOfJobs; i++ {
		log.Println("Waiting to launch", i)

		<-semaphore
		log.Println("It's my turn", i)

		go func(i int) {
			asyncWork()
			log.Println("Done!", i)
			done <- true
		}(i)
	}
	<-waitAll
}
