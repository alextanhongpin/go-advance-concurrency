package main

import (
	"log"
	"time"
)

var c = make(chan int, 5)

func main() {

	go worker()

	for i := 0; i < 10; i++ {
		c <- i
		log.Println(i)
	}

}

func worker() {
	for {
		_ = <-c
		time.Sleep(time.Second)
	}
}
