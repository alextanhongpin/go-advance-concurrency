# Backpressure

## Status

`inactive`

## Context

I can't remember what this example is meant to demonstrate.

```go
package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const batchTime = 3

func init() {
	rand.Seed(time.Now().UnixNano())
}

type RunContext struct {
	inChan    chan Event
	batchChan chan EventBatch
	backChan  chan EventBatch
	doneChan  chan bool

	// For reporting purpose
	sumProduced int64
	sumSent     int64
}

// Event simple event
type Event int

// Slice of events
type EventBatch []Event

func NewRunContext() *RunContext {
	return &RunContext{
		inChan:    make(chan Event),
		batchChan: make(chan EventBatch),
		backChan:  make(chan EventBatch, 1),
		doneChan:  make(chan bool),
	}
}

func NewEventBatch() EventBatch {
	return EventBatch{}
}

func (e EventBatch) Merge(other Event) EventBatch {
	return append(e, other)
}

func produce(runCtx *RunContext) {
	defer close(runCtx.inChan)
	for {
		delay := time.Duration(rand.Intn(5)+1) * time.Second
		time.Sleep(delay)

		nMessages := rand.Intn(10) + 1

		for i := 0; i < nMessages; i++ {
			// Generate a random value
			e := Event(rand.Intn(10))

			fmt.Println("Producing:", e)
			select {
			case runCtx.inChan <- e:
				atomic.AddInt64(&runCtx.sumProduced, int64(e))
			case <-runCtx.doneChan:
				fmt.Println("producer completed")
				return
			}
		}
	}
}

func run(runCtx *RunContext, wg *sync.WaitGroup) {
	defer wg.Done()

	eventBatch := NewEventBatch()

	ticker := time.Tick(time.Duration(batchTime) * time.Second)

LOOP:
	for {
		select {
		case ev, ok := <-runCtx.inChan:
			if !ok {
				if len(eventBatch) > 0 {
					fmt.Println("Dispatching last batch")

					runCtx.batchChan <- eventBatch
				}
				close(runCtx.batchChan)
				fmt.Println("run finished")
				break LOOP
			}
			eventBatch = eventBatch.Merge(ev)
		case <-ticker:
			if len(eventBatch) > 0 {
				fmt.Println("waiting to send")
				runCtx.batchChan <- eventBatch
				eventBatch = <-runCtx.backChan
			}
		}

	}
}

func batchWriter(runCtx *RunContext, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		eb, ok := <-runCtx.batchChan
		if !ok {
			fmt.Println("batch writer completed")
			return
		}

		// Simulate time to persist the batch using a random delay
		delay := time.Duration(rand.Intn(3)+1) * time.Second
		time.Sleep(delay)

		// Perform retry if  this write fails and add an exponential backoff

		for _, e := range eb {
			atomic.AddInt64(&runCtx.sumSent, int64(e))
		}

		fmt.Println("batch sent:", eb, delay)
		runCtx.backChan <- NewEventBatch()
	}
}

func waitOnSignal(runCtx *RunContext, sigs <-chan os.Signal) {
	fmt.Println("awaiting signals")
	sig := <-sigs
	fmt.Println(sig)
	// Shut down input
	close(runCtx.doneChan)
}

func main() {

	var wg sync.WaitGroup

	runCtx := NewRunContext()
	sigs := make(chan os.Signal, 1)

	// if you hit CTRL-C or kill the process this channel will
	// get a signal and trigger a shutdown of the publisher
	// which in turn should trigger a each step of the pipeline
	// to exit
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go waitOnSignal(runCtx, sigs)
	go produce(runCtx)

	wg.Add(2)
	go run(runCtx, &wg)
	go batchWriter(runCtx, &wg)

	wg.Wait()

	fmt.Println("summary")
	fmt.Printf("	produced: %d\n", runCtx.sumProduced)
	fmt.Printf("	sent: %d\n", runCtx.sumSent)
}

```
