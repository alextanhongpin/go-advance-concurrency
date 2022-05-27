# Scheduling parallel tasks with future

```go
// You can edit this code!
// Click here and start typing.
package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func parallelTask() error {
	fmt.Println("is parallel")
	// return task()
	return errors.New("failed at parallel task")
}
func task() error {
	n := rand.Intn(10)
	fmt.Println("doing work", n)
	time.Sleep(1 * time.Second)
	fmt.Println("done", n)
	return nil
}

func main() {
	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()

	parallelTask := func() chan error {
		err := <-Future(parallelTask)
		if err != nil {
			ch := make(chan error, 1)
			ch <- err
			close(ch)

			return ch
		}
		return All(Future(parallelTask), Future(parallelTask))
	}

	errCh := All(Future(task), Future(task), Future(task), parallelTask())
	for err := range errCh {
		fmt.Println("received", err)
	}
}

type Task func() error

func All(chans ...chan error) chan error {
	errCh := make(chan error)

	var wg sync.WaitGroup
	wg.Add(len(chans))

	for _, ch := range chans {
		go func(ch chan error) {
			defer wg.Done()
			for err := range ch {
				errCh <- err
			}

		}(ch)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	return errCh
}

func Future(task Task) chan error {
	ch := make(chan error)

	go func() {
		defer close(ch)
		ch <- task()
	}()

	return ch
}
```
