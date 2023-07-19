# Backpressure

## Status

`proposed`


## Context

We want to apply backpressure, and drop requests when we can no longer handle incoming requests.


## Decision

We will implement backpressure as a middleware. The backpressure can be applied at selected paths after determining which endpoints are more prone to high load.


## Consequences


The system should be able to shed some load and handle request without failing.

However, requests will be dropped for some clients until the system is able to serve them.

```go
 package main

 import (
     "fmt"
     "math/rand"
     "sync"
     "time"
 )

 func main() {
     // Increase the number of worker to process more items.
     numWorker := 1

     // Create a queue with max buffer of 10 items.
     ch := make(chan int, 10)

     // Our SLA is 10 millisecond of processing time.
     // If the queue is full, and the SLA is violated, we drop the event and notify the user.
     sla := 25 * time.Millisecond

     send := func(n int) bool {
         select {
         case ch <- n:
             return true
         case <-time.After(sla):
             return false
         }
     }

     var wg sync.WaitGroup
     for i := 0; i < numWorker; i++ {
         wg.Add(1)

         go func(i int) {
             defer wg.Done()

             var count int
             // Simulate a slow consumer
             for range ch {
                 sleep := time.Duration(rand.Intn(50)) * time.Millisecond
                 time.Sleep(sleep)
                 count++
             }

             fmt.Println("recv:", count, "worker:", i)
         }(i)
     }

     n := 50
     fmt.Println("send:", n)
     for i := 0; i < n; i++ {
         // TODO: Notify user that their request will not be handled.
         if !send(i) {
             fmt.Println("drop:", i)
         }
     }
     close(ch)
     wg.Wait()

     fmt.Println("program exiting")
 }
```
