package main

import "time"

func main() {
	rate := time.Second / 10
	burstLimit := 100
	tick := time.NewTicker(rate)

	defer tick.Stop()

	throttle := make(chan time.Time, burstLimit)
	go func (){
		for t := range tick.C {
			select {
			case throttle <- t:
				default
			}
		}
	}()

	for req := range requests {
		<- throttle
		go client.Call("Service.Method", req)
	}
}