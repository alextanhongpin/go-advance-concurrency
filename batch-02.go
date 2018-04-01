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
)

func main() {

	rand.Seed(time.Now().UnixNano())

	generator := func(done <-chan interface{}, limit int) <-chan interface{} {
		outStream := make(chan interface{})

		go func() {
			defer close(outStream)
			for i := 0; i < limit; i++ {
				time.Sleep(time.Duration(rand.Intn(10)+10) * time.Millisecond)
				outStream <- i
			}
		}()

		return outStream
	}

	done := make(chan interface{})
	defer close(done)

	batch := []int{}
	out := make(chan interface{})

	go func() {
		inStream := generator(done, N_ITEMS)
		for {
			select {
			case <-done:
				if len(batch) > 0 {
					log.Println("done, processing remaining", len(batch))
					for _, v := range batch {
						out <- v
					}
					batch = []int{}
					close(out)
				}
			case <-time.After(20 * time.Millisecond):
				log.Println("timeout exceeded, processing", len(batch))
				if len(batch) > 0 {
					for _, v := range batch {
						out <- v
					}
					batch = []int{}
				}
			case v, ok := <-inStream:
				if !ok {
					log.Println("channel closed, processing remaining", len(batch))
					if len(batch) > 0 {
						for _, v := range batch {
							out <- v
						}
						batch = []int{}
					}
					close(out)
					return
				}
				batch = append(batch, v.(int))
				if len(batch) >= MAX_BATCH_SIZE {
					log.Println("threshold exceeded, processing", len(batch))
					for _, i := range batch {
						out <- i
					}
					batch = []int{}
				}
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(N_ITEMS)
	go func() {
		for o := range out {
			defer wg.Done()
			log.Println("got", o)
		}
	}()
	wg.Wait()
	log.Println("done")
}

// 2018/03/26 15:07:30 threshold exceeded, processing 10
// 2018/03/26 15:07:30 got 0
// 2018/03/26 15:07:30 got 1
// 2018/03/26 15:07:30 got 2
// 2018/03/26 15:07:30 got 3
// 2018/03/26 15:07:30 got 4
// 2018/03/26 15:07:30 got 5
// 2018/03/26 15:07:30 got 6
// 2018/03/26 15:07:30 got 7
// 2018/03/26 15:07:30 got 8
// 2018/03/26 15:07:30 got 9
// 2018/03/26 15:07:30 threshold exceeded, processing 10
// 2018/03/26 15:07:30 got 10
// 2018/03/26 15:07:30 got 11
// 2018/03/26 15:07:30 got 12
// 2018/03/26 15:07:30 got 13
// 2018/03/26 15:07:30 got 14
// 2018/03/26 15:07:30 got 15
// 2018/03/26 15:07:30 got 16
// 2018/03/26 15:07:30 got 17
// 2018/03/26 15:07:30 got 18
// 2018/03/26 15:07:30 got 19
// 2018/03/26 15:07:30 timeout exceeded, processing 0
// 2018/03/26 15:07:30 timeout exceeded, processing 9
// 2018/03/26 15:07:30 got 20
// 2018/03/26 15:07:30 got 21
// 2018/03/26 15:07:30 got 22
// 2018/03/26 15:07:30 got 23
// 2018/03/26 15:07:30 got 24
// 2018/03/26 15:07:30 got 25
// 2018/03/26 15:07:30 got 26
// 2018/03/26 15:07:30 got 27
// 2018/03/26 15:07:30 got 28
// 2018/03/26 15:07:30 timeout exceeded, processing 8
// 2018/03/26 15:07:30 got 29
// 2018/03/26 15:07:30 got 30
// 2018/03/26 15:07:30 got 31
// 2018/03/26 15:07:30 got 32
// 2018/03/26 15:07:30 got 33
// 2018/03/26 15:07:30 got 34
// 2018/03/26 15:07:30 got 35
// 2018/03/26 15:07:30 got 36
// 2018/03/26 15:07:31 threshold exceeded, processing 10
// 2018/03/26 15:07:31 got 37
// 2018/03/26 15:07:31 got 38
// 2018/03/26 15:07:31 got 39
// 2018/03/26 15:07:31 got 40
// 2018/03/26 15:07:31 got 41
// 2018/03/26 15:07:31 got 42
// 2018/03/26 15:07:31 got 43
// 2018/03/26 15:07:31 got 44
// 2018/03/26 15:07:31 got 45
// 2018/03/26 15:07:31 got 46
// 2018/03/26 15:07:31 threshold exceeded, processing 10
// 2018/03/26 15:07:31 got 47
// 2018/03/26 15:07:31 got 48
// 2018/03/26 15:07:31 got 49
// 2018/03/26 15:07:31 got 50
// 2018/03/26 15:07:31 got 51
// 2018/03/26 15:07:31 got 52
// 2018/03/26 15:07:31 got 53
// 2018/03/26 15:07:31 got 54
// 2018/03/26 15:07:31 got 55
// 2018/03/26 15:07:31 got 56
// 2018/03/26 15:07:31 threshold exceeded, processing 10
// 2018/03/26 15:07:31 got 57
// 2018/03/26 15:07:31 got 58
// 2018/03/26 15:07:31 got 59
// 2018/03/26 15:07:31 got 60
// 2018/03/26 15:07:31 got 61
// 2018/03/26 15:07:31 got 62
// 2018/03/26 15:07:31 got 63
// 2018/03/26 15:07:31 got 64
// 2018/03/26 15:07:31 got 65
// 2018/03/26 15:07:31 got 66
// 2018/03/26 15:07:31 threshold exceeded, processing 10
// 2018/03/26 15:07:31 got 67
// 2018/03/26 15:07:31 got 68
// 2018/03/26 15:07:31 got 69
// 2018/03/26 15:07:31 got 70
// 2018/03/26 15:07:31 got 71
// 2018/03/26 15:07:31 got 72
// 2018/03/26 15:07:31 got 73
// 2018/03/26 15:07:31 got 74
// 2018/03/26 15:07:31 got 75
// 2018/03/26 15:07:31 got 76
// 2018/03/26 15:07:31 threshold exceeded, processing 10
// 2018/03/26 15:07:31 got 77
// 2018/03/26 15:07:31 got 78
// 2018/03/26 15:07:31 got 79
// 2018/03/26 15:07:31 got 80
// 2018/03/26 15:07:31 got 81
// 2018/03/26 15:07:31 got 82
// 2018/03/26 15:07:31 got 83
// 2018/03/26 15:07:31 got 84
// 2018/03/26 15:07:31 got 85
// 2018/03/26 15:07:31 got 86
// 2018/03/26 15:07:31 threshold exceeded, processing 10
// 2018/03/26 15:07:31 got 87
// 2018/03/26 15:07:31 got 88
// 2018/03/26 15:07:31 got 89
// 2018/03/26 15:07:31 got 90
// 2018/03/26 15:07:31 got 91
// 2018/03/26 15:07:31 got 92
// 2018/03/26 15:07:31 got 93
// 2018/03/26 15:07:31 got 94
// 2018/03/26 15:07:31 got 95
// 2018/03/26 15:07:31 got 96
// 2018/03/26 15:07:31 channel closed, processing remaining 3
// 2018/03/26 15:07:31 got 97
// 2018/03/26 15:07:31 got 98
// 2018/03/26 15:07:31 got 99
// 2018/03/26 15:07:31 done
