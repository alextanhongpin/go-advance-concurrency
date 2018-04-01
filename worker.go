package main

import (
	"log"
	"time"
)

func main() {
	concurrency := 5
	sem := make(chan bool, concurrency)
	urls := []string{"url1", "url2", "url3", "url4", "url5", "url6", "url7", "url8"}

	for _, url := range urls {
		sem <- true
		log.Println("add sem:", len(sem))
		go func(url string) {
			defer func() {
				<-sem
				log.Println("clear sem:", len(sem))
			}()
			time.Sleep(2 * time.Second)
			log.Println("Handling url", url)
		}(url)
	}
	log.Println("capicity", cap(sem), len(sem))
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
}
