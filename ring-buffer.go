// This program demonstrates how to use ring buffer
package main

import (
	"fmt"
)

type RingBuffer struct {
	inputChannel  <-chan int
	outputChannel chan int
}

func NewRingBuffer(inputChannel <-chan int, outputChannel chan int) *RingBuffer {
	return &RingBuffer{inputChannel, outputChannel}
}

func (r *RingBuffer) Run() {
	for v := range r.inputChannel {
		select {
		case r.outputChannel <- v:
		default:
			// Slower consumer might lose (oldest) messages, but will never be able to block the main message
			// processing queue
			fmt.Println("current value", v)
			fmt.Println("in queue value", <-r.outputChannel)
			r.outputChannel <- v
		}
	}
	close(r.outputChannel)
}

func main() {
	in := make(chan int)
	out := make(chan int, 5)
	rb := NewRingBuffer(in, out)
	go rb.Run()

	for i := 0; i < 10; i++ {
		in <- i
	}

	close(in)

	for res := range out {
		fmt.Println(res)
	}
}

// current value 5
// in queue value 0
// current value 6
// in queue value 1
// current value 7
// in queue value 2
// current value 8
// in queue value 3
// 4
// 5
// 6
// 7
// 8
// 9
