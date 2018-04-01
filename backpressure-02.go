package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	requests := make(chan request, 100)
	go startServer(requests)
	go process(requests)

	makeRequests(500, 6*time.Millisecond)
	log.Println("second round")
	time.Sleep(1 * time.Second)
	makeRequests(500, 6*time.Millisecond)
}

func startServer(req chan request) {
	http.HandleFunc("/requests", handle(req))
	http.ListenAndServe(":9000", nil)
}

func process(req chan request) {
	for r := range req {
		r.process()
	}
}

func makeRequests(count int, cooldown time.Duration) {
	wg := sync.WaitGroup{}
	for i := 0; i < count; i++ {
		go func() {
			wg.Add(1)
			defer wg.Done()

			response, err := http.Get("http://localhost:9000/requests")
			if err != nil {
				log.Println(err)
				return
			}
			defer response.Body.Close()

			b, _ := ioutil.ReadAll(response.Body)
			fmt.Println(string(b))

		}()
		time.Sleep(cooldown)
	}
	wg.Wait()
}

func handle(req chan request) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(req) < cap(req) {
			r := newRequest(r)
			req <- r
			w.Write(<-r.response)
		} else {
			w.Write([]byte("x"))
		}
	}
}

type request struct {
	r        *http.Request
	response chan []byte
}

func newRequest(r *http.Request) request {
	return request{
		r, make(chan []byte),
	}
}

func (r request) process() {
	time.Sleep(10 * time.Millisecond)
	r.response <- []byte("o")
}
