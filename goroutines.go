package main

import (
	"log"
	"runtime/pprof"
	"time"
)

func main() {

	go func() {
		time.Sleep(1 * time.Second)
		log.Println("hello 1")
	}()

	go func() {
		goroutines := pprof.Lookup("goroutine")
		for range time.Tick(1 * time.Second) {
			log.Printf("goroutine count: %d\n", goroutines.Count())
		}
	}()

	time.Sleep(10 * time.Second)
}
