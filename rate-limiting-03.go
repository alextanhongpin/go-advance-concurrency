package main

import "time"

func main() {
	// works for rates up to tens per second
	// for higher limit, use token bucket rate
	rate := time.Second / 10
	throttle := time.Tick(rate)

	for req := range requests {
		<-throttle
		go client.Call("service.Method", req)
	}
}
