// This program demonstrates how to write a web crawler with a single worker and multiple concurrent workers
package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type Response struct {
	Body  string
	Error error
}

func fetch(url string) *Response {
	resp, _ := http.Get(url)
	body, err := ioutil.ReadAll(resp.Body)

	return &Response{
		Body:  string(body),
		Error: err,
	}
}

func main() {
	urls := []string{
		"http://www.mastergoco.com/index1.html",
		"http://www.mastergoco.com/index2.html",
		"http://www.mastergoco.com/index3.html",
		"http://www.mastergoco.com/index4.html",
		"http://www.mastergoco.com/index5.html",
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Crawl pages using only one worker
	start0 := time.Now()
	for o := range singleWorker(ctx, urls...) {
		res := o.(*Response)
		if res.Error != nil {
			log.Println("error at single worker:", res.Error)
			continue
		}
		log.Println("single worker done", len(res.Body))
	}
	log.Println("single worker completed in", time.Since(start0))

	log.Println("")
	// Crawl pages using multiple workers
	numWorkers := 4
	start1 := time.Now()
	for o := range multiWorkers(ctx, numWorkers, toStream(ctx, urls...)) {
		res := o.(*Response)
		if res.Error != nil {
			log.Println("error at multi worker:", res.Error)
			continue
		}
		log.Println("multi worker done", len(res.Body))
	}
	log.Println("multi workers completed in", time.Since(start1))
}

func singleWorker(ctx context.Context, urls ...string) <-chan interface{} {
	outStream := make(chan interface{})

	go func() {
		defer close(outStream)

		for _, url := range urls {
			outStream <- fetch(url)
		}
	}()

	return outStream
}

func multiWorkers(ctx context.Context, numWorkers int, inStream <-chan interface{}) <-chan interface{} {
	outStream := make(chan interface{})

	var wg sync.WaitGroup
	wg.Add(numWorkers)
	multiplex := func(index int, in <-chan interface{}) {
		for o := range in {
			select {
			case <-ctx.Done():
				return
			case outStream <- fetch(o.(string)):
				log.Println("processed by worker", index)
			}
		}
		wg.Done()
	}

	for i := 0; i < numWorkers; i++ {
		go multiplex(i, inStream)
	}

	go func() {
		wg.Wait()
		close(outStream)
	}()

	return outStream
}

func toStream(ctx context.Context, urls ...string) <-chan interface{} {
	outStream := make(chan interface{})

	go func() {
		for _, v := range urls {
			select {
			case <-ctx.Done():
				return
			case outStream <- v:
			}
		}
		close(outStream)
	}()

	return outStream
}

// 2018/03/29 11:47:50 single worker done 5943
// 2018/03/29 11:47:50 single worker done 5943
// 2018/03/29 11:47:51 single worker done 5943
// 2018/03/29 11:47:51 single worker done 5943
// 2018/03/29 11:47:52 single worker done 5943
// 2018/03/29 11:47:52 single worker completed in 3.005570534s
// 2018/03/29 11:47:52
// 2018/03/29 11:47:53 processed by worker 3
// 2018/03/29 11:47:53 multi worker done 5943
// 2018/03/29 11:47:53 processed by worker 1
// 2018/03/29 11:47:53 multi worker done 5943
// 2018/03/29 11:47:53 processed by worker 0
// 2018/03/29 11:47:53 multi worker done 5943
// 2018/03/29 11:47:53 processed by worker 2
// 2018/03/29 11:47:53 multi worker done 5943
// 2018/03/29 11:47:53 processed by worker 3
// 2018/03/29 11:47:53 multi worker done 5943
// 2018/03/29 11:47:53 multi workers completed in 1.140311726s
