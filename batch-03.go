package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	N_ITEMS        = 100
	MAX_BATCH_SIZE = 10
	MAX_DURATION   = 20
)

type Batch struct {
	MaxBatchSize int
	Queue        []int
	Duration     time.Duration
	Out          chan []int
}

func (b *Batch) Process() {
	if !b.HasItems() {
		return
	}
	// Send the whole batch to the channel
	b.Out <- b.Queue

	// Empty queue
	b.Queue = []int{}
}

func (b *Batch) IsFull() bool {
	return len(b.Queue) == b.MaxBatchSize
}

func (b *Batch) HasItems() bool {
	return len(b.Queue) > 0
}
func (b *Batch) Size() int {
	return len(b.Queue)
}

func (b *Batch) Close() {
	close(b.Out)
}

func (b *Batch) Run(done, in <-chan interface{}) {
	go func() {
		defer b.Close()
		for {
			select {
			case <-done:
				log.Println("done, processing remaining", b.Size())
				b.Process()
			case <-time.After(b.Duration):
				log.Println("timeout exceeded, processing", b.Size())
				b.Process()
			case v, ok := <-in:
				if !ok {
					log.Println("channel closed, processing remaining", b.Size())
					b.Process()
					return
				}
				b.Queue = append(b.Queue, v.(int))
				if b.IsFull() {
					log.Println("threshold exceeded, processing", b.Size())
					b.Process()
					// We can add a delay here to allow some buffer time for processing
					// log.Println("sleep 1s before next batch")
					// time.Sleep(1 * time.Second)
				}
			}
		}
	}()
}

func NewBatch(maxBatchSize, duration int) *Batch {
	return &Batch{
		MaxBatchSize: maxBatchSize,
		Queue:        []int{},
		Duration:     time.Duration(duration) * time.Millisecond,
		Out:          make(chan []int),
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	generator := func(done <-chan interface{}, limit int) <-chan interface{} {
		outStream := make(chan interface{})

		go func() {
			defer close(outStream)
			for i := 0; i < limit; i++ {
				time.Sleep(time.Duration(rand.Intn(10)+10) * time.Millisecond)
				select {
				case <-done:
					return
				case outStream <- i:
				}
			}
		}()

		return outStream
	}

	done := make(chan interface{})
	defer close(done)

	in := generator(done, N_ITEMS)

	batch := NewBatch(MAX_BATCH_SIZE, MAX_DURATION)
	batch.Run(done, in)

	var wg sync.WaitGroup
	wg.Add(N_ITEMS)
	go func() {
		var sum int
		for o := range batch.Out {
			log.Println("got", o)
			for _ = range o {
				defer wg.Done()
				// log.Println("each", v)
				sum++
			}
		}
		log.Printf("processed %d items\n", sum)
	}()
	wg.Wait()
	log.Println("done")
}

// 2018/03/26 22:52:42 timeout exceeded, processing 3
// 2018/03/26 22:52:42 got [0 1 2]
// 2018/03/26 22:52:43 threshold exceeded, processing 10
// 2018/03/26 22:52:43 got [3 4 5 6 7 8 9 10 11 12]
// 2018/03/26 22:52:43 threshold exceeded, processing 10
// 2018/03/26 22:52:43 got [13 14 15 16 17 18 19 20 21 22]
// 2018/03/26 22:52:43 threshold exceeded, processing 10
// 2018/03/26 22:52:43 got [23 24 25 26 27 28 29 30 31 32]
// 2018/03/26 22:52:43 timeout exceeded, processing 8
// 2018/03/26 22:52:43 got [33 34 35 36 37 38 39 40]
// 2018/03/26 22:52:43 timeout exceeded, processing 2
// 2018/03/26 22:52:43 got [41 42]
// 2018/03/26 22:52:43 threshold exceeded, processing 10
// 2018/03/26 22:52:43 got [43 44 45 46 47 48 49 50 51 52]
// 2018/03/26 22:52:43 threshold exceeded, processing 10
// 2018/03/26 22:52:43 got [53 54 55 56 57 58 59 60 61 62]
// 2018/03/26 22:52:43 timeout exceeded, processing 4
// 2018/03/26 22:52:43 got [63 64 65 66]
// 2018/03/26 22:52:44 threshold exceeded, processing 10
// 2018/03/26 22:52:44 got [67 68 69 70 71 72 73 74 75 76]
// 2018/03/26 22:52:44 threshold exceeded, processing 10
// 2018/03/26 22:52:44 got [77 78 79 80 81 82 83 84 85 86]
// 2018/03/26 22:52:44 threshold exceeded, processing 10
// 2018/03/26 22:52:44 got [87 88 89 90 91 92 93 94 95 96]
// 2018/03/26 22:52:44 channel closed, processing remaining 3
// 2018/03/26 22:52:44 got [97 98 99]
// 2018/03/26 22:52:44 processed 100 items
// 2018/03/26 22:52:44 done
