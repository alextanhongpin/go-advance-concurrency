package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func main() {
	newRandStream := func(ctx context.Context) <-chan int {
		randStream := make(chan int)
		go func() {
			defer fmt.Println("newRandStream closure exited")
			defer close(randStream)
			for {
				select {
				case <-ctx.Done():
					return
				case randStream <- rand.Int():
				}
			}
		}()
		return randStream
	}

	ctx, cancel := context.WithCancel(context.Background())
	randStream := newRandStream(ctx)
	for i := 1; i <= 3; i++ {
		fmt.Printf("%d: %d\n", i, <-randStream)
	}
	cancel()
	time.Sleep(1 * time.Second)
}
