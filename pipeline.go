package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// StreamResult holds the value and err as it goes through multiple pipelines
type StreamResult struct {
	err   error
	value int
}

// THRESHOLD is the maximum value that is allowed
const THRESHOLD = 50

func main() {

	source := func(ctx context.Context) <-chan StreamResult {
		intStream := make(chan StreamResult)
		go func() {
			for {
				result := StreamResult{
					value: rand.Intn(100),
				}

				if result.value > THRESHOLD {
					result.err = errors.New("exceeded threshold")
					result.value = -1
				}

				select {
				case <-ctx.Done():
					return
				case intStream <- result:
				}
			}
		}()
		return intStream
	}

	processor := func(ctx context.Context, intStream <-chan StreamResult) <-chan StreamResult {
		multiplierStream := make(chan StreamResult)
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case res := <-intStream:
					if res.err != nil {
						multiplierStream <- res
						continue
					}
					res.value *= 100
					multiplierStream <- res
				}
			}
		}()
		return multiplierStream
	}

	sink := func(source <-chan StreamResult, upperBoundary int, thresholdPercentage float64) {
		// SLA, only 1% error can be tolerated
		maxAllowedError := int(math.Ceil(float64(upperBoundary) * thresholdPercentage))
		count := 0

		for i := 0; i < upperBoundary; i++ {
			result := <-source

			// Handle error silently
			if result.err != nil {
				count++
				if count > maxAllowedError {
					fmt.Printf("error threshold exceeded: %d, completed: %d/%d\n", count, i, upperBoundary)
					break
				}
				fmt.Printf("tolerated error: %s\n", result.err.Error())
				continue
			}

			// Print out the value for debugging
			fmt.Printf("got %d, completed: %d/%d\n", result.value, i, upperBoundary)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	// source will generate random integer in range 0-100, but treat those below 50 as errors
	p1 := source(ctx)

	// Multiply valid values (>50) by 100
	p2 := processor(ctx, p1)
	sink(p2, 100, 0.25)

	// Terminate all the pipeline
	cancel()

	time.Sleep(1 * time.Second)
	fmt.Println("done")
}
