package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

type request struct {
	r  *http.Request
	ch chan []byte
}

func newRequest(r *http.Request) request {
	return request{
		r:  r,
		ch: make(chan []byte),
	}
}

func main() {
	// We create a buffered channel with a max capacity of 1000 items.
	// If the queue is full, drop the requests.
	q := make(chan request, 100)
	h := func(w http.ResponseWriter, r *http.Request) {
		if len(q) < cap(q) {
			req := newRequest(r) // Process the request.
			q <- req             // Send to queue.
			w.Write(<-req.ch)    // Wait for the reply.
		} else {
			w.Write([]byte("x")) // Drop the request.
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(h))
	defer ts.Close()

	// Process the request in the background.
	go process(q)

	makeRequests(ts.URL, 500, 6*time.Millisecond)
	log.Println("second round")
	time.Sleep(1 * time.Second)
	makeRequests(ts.URL, 500, 6*time.Millisecond)
}

func process(q chan request) {
	for r := range q {
		// Sleep to simulate delay.
		time.Sleep(10 * time.Millisecond)

		// Send resp.
		r.ch <- []byte("o")
	}
}

func makeRequests(url string, count int, cooldown time.Duration) {
	wg := sync.WaitGroup{}
	for i := 0; i < count; i++ {
		go func() {
			wg.Add(1)
			defer wg.Done()

			resp, err := http.Get(url)
			if err != nil {
				log.Println(err)
				return
			}
			defer resp.Body.Close()

			b, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(string(b))
		}()
		time.Sleep(cooldown)
	}
	wg.Wait()
}
