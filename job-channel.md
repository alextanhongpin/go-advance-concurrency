## Job Channel

Create a struct with result channel and payload, and pass them to another channel to be processed. Upon completion, the result is send back to the result channel.

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type Job struct {
	resultCh chan string
	payload  string
}

func main() {

	worker := make(chan Job)
	numWorkers := 3
	for i := 0; i < numWorkers; i++ {
		go func() {
			for job := range worker {
				time.Sleep(1 * time.Second)
				job.resultCh <-job.payload + " completed"
			}
		}()
	}

	n := 10

	var wg sync.WaitGroup
	wg.Add(n)

	doWork := func(i int) {
		defer wg.Done()

		job := Job{
			resultCh: make(chan string), 
			payload: fmt.Sprintf("job#%d", i),
		}

		worker <- job
		msg := <-job.resultCh
		fmt.Println(msg)
	}

	for i := 0; i < n; i++ {
		go func(i int) {
			doWork(i)
		}(i)
	}

	wg.Wait()
	fmt.Println("exiting")
}
```
