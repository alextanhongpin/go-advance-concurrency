package main

import (
	"log"
	"time"
)

func main() {
	// limit := make(chan int, 5)
	ch := make(chan int, 10)

	go func() {
		for {
			select {
			case x := <-ch:
				log.Println(x)
			default:
				log.Println("full")
				log.Println(<-ch)
			}
		}
	}()

	for i := 0; i < 1000; i++ {
		ch <- i
	}

	time.Sleep(5 * time.Second)
}
