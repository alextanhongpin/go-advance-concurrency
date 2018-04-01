package main

import (
	"log"
	"time"
)

func main() {

	maxConcurrency := 3

	// We set the max concurrency to 3 - which means me limit the goroutine to run
	// a maximum of 3 at a time
	semaphore := make(chan bool, maxConcurrency)
	urls := []string{"url1", "url2", "url3", "url4", "url5", "url6", "url7", "url8", "url9", "url10"}

	for i, url := range urls {
		log.Println("Starting", i)
		semaphore <- true
		log.Println("Semaphore", cap(semaphore), len(semaphore))
		go func(u string) {
			log.Println("Doing something with url", u)
			defer func() {
				<-semaphore
			}()
			time.Sleep(1 * time.Second)
			log.Println("Got this url", u)
		}(url)
		log.Println("Ending", i)
	}
	log.Println("Loop done")

	log.Println("Clearing...")
	for i := 0; i < cap(semaphore); i++ {
		semaphore <- true
		log.Println("Clearing sem", i)
	}
	log.Println("Exiting...")
}
