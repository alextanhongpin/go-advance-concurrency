package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	wg sync.WaitGroup
)

func work() error {
	defer wg.Done()

	for i := 0; i < 1000; i++ {
		select {
		case <-time.After(2 * time.Second):
			fmt.Println("Doing some work", i)
		}
	}
	return nil
}

func main() {
	fmt.Println("Hey, I'm going to do some work")

	ch := make(chan error, 1)

	go func() {
		ch <- work()
	}()

	select {
	case err := <-ch:
		log.Fatal("Something went wrong", err)
	case <-time.After(4 * time.Second):
		fmt.Println("Life is short to wait that long")
	}

	fmt.Println("Done")
}
