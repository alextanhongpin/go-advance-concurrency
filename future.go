package main

import (
	"log"
	"time"
)

func main() {

	futureDouble := func(val int) <-chan interface{} {
		outStream := make(chan interface{})
		go func() {
			defer close(outStream)
			// Long running process
			time.Sleep(1 * time.Second)
			log.Println("computed future")
			outStream <- val * 2
		}()
		return outStream
	}

	resp := futureDouble(100)

	log.Println("do something...")
	time.Sleep(2 * time.Second)
	log.Println("do something else...")
	time.Sleep(2 * time.Second)
	log.Println("getting future result")

	log.Println(<-resp)
}
