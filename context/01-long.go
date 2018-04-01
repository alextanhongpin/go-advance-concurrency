package main

import (
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup

func work() error {

	defer wg.Done()
	for i := 0; i < 1000; i++ {
		select {
		case <-time.After(2 * time.Second):
			fmt.Println("Doing work", i)
		}
	}
	return nil
}

func main() {
	fmt.Println("Hey. I'm going to do some work.")

	wg.Add(1)
	go work()
	wg.Wait()

	fmt.Println("Done")
}
