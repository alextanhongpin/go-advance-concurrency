## Heartbeat

```go
package main

import (
	"fmt"
	"time"
)

func doWork(
	done chan interface{},
	pulseInterval time.Duration,
) (<-chan interface{}, <-chan time.Time) {
	heartbeat := make(chan interface{})
	results := make(chan time.Time)

	go func() {
		defer close(heartbeat)
		defer close(results)

		pulse := time.Tick(pulseInterval)
		workGen := time.Tick(2 * pulseInterval)
		sendPulse := func() {
			select {
			case heartbeat <- struct{}{}:
			default:
			}
		}
		sendResult := func(r time.Time) {
			for {
				select {
				case <-done:
					return
				case <-pulse:
					sendPulse()
				case results <- r:
					return
				}
			}
		}

		for {
			select {
			case <-done:
				return
			case <-pulse:
				sendPulse()
			case r := <-workGen:
				sendResult(r)
			}
		}
	}()
	return heartbeat, results
}

func main() {
	done := make(chan interface{})
	time.AfterFunc(10*time.Second, func() {
		close(done)
	})
	const timeout = 2 * time.Second
	heartbeat, results := doWork(done, timeout/2)
	for {
		select {
		case _, ok := <-heartbeat:
			if !ok {
				return
			}
			fmt.Println("pulse")
		case r, ok := <-results:
			if !ok {
				return
			}
			fmt.Println("results", r.Second())
		case <-time.After(timeout):
			return
		}
	}
}
```

## Unhealthy Goroutine

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	doWork := func(
		done <-chan interface{},
		pulseInterval time.Duration,
	) (<-chan interface{}, <-chan time.Time) {
		heartbeat := make(chan interface{})
		results := make(chan time.Time)

		go func() {
			// defer close(heartbeat)
			// defer close(results)

			pulse := time.Tick(pulseInterval)
			workGen := time.Tick(2 * pulseInterval)
			sendPulse := func() {
				select {
				case heartbeat <- struct{}{}:
				default:
				}
			}
			sendResult := func(r time.Time) {
				for {
					select {
					case <-pulse:
						sendPulse()
					case results <- r:
						return
					}
				}
			}

			for i := 0; i < 2; i++ {
				select {
				case <-done:
					return
				case <-pulse:
					sendPulse()
				case r := <-workGen:
					sendResult(r)
				}
			}
		}()
		return heartbeat, results
	}

	done := make(chan interface{})
	time.AfterFunc(10*time.Second, func() {
		close(done)
	})
	const timeout = 2 * time.Second
	heartbeat, results := doWork(done, timeout/2)
	for {
		select {
		case _, ok := <-heartbeat:
			if !ok {
				return
			}
			fmt.Println("pulse")
		case r, ok := <-results:
			if !ok {
				return
			}
			fmt.Printf("results: %v\n", r.Second())
		case <-time.After(timeout):
			fmt.Println("worker goroutine is not healthy")
			return
		}
	}
}
```

## Sending pulse before every work

```go
package main

import (
	"fmt"
)

func main() {
	doWork := func(
		done <-chan interface{},
	) (<-chan interface{}, <-chan int) {
		heartbeatStream := make(chan interface{}, 1)
		workStream := make(chan int)
		go func() {
			defer close(heartbeatStream)
			defer close(workStream)

			for i := 0; i < 10; i++ {
				select {
				case heartbeatStream <- struct{}{}:
				default:
				}

				select {
				case <-done:
					return
				case workStream <- i:
				}
			}
		}()
		return heartbeatStream, workStream
	}

	done := make(chan interface{})
	defer close(done)

	heartbeat, results := doWork(done)
	for {
		select {
		case _, ok := <-heartbeat:
			if !ok {
				return
			}
			fmt.Println("pulse")
		case r, ok := <-results:
			if !ok {
				return
			}
			fmt.Println("results", r)
		}
	}
}
```

## Heartbeat at fixed interval

```go
package main

import (
	"fmt"
	"time"
)

type Builder interface {
	SetContent()
}

func main() {

	doWork := func(done chan interface{}, in <-chan interface{}) <-chan interface{} {
		outStream := make(chan interface{})
		ticker := time.NewTicker(1 * time.Second)
		go func() {
			defer close(outStream)
			defer ticker.Stop()
			for {
				select {
				case <-done:
					return
				case t := <-ticker.C:
					fmt.Println("heartbeat", t)
				default:
				}

				select {
				case <-done:
					return
				case i, ok := <-in:
					if !ok {
						return
					}
					outStream <- i
				case <-done:
					return
				}
			}
		}()
		return outStream
	}

	generator := func(done chan interface{}, values ...interface{}) <-chan interface{} {
		outStream := make(chan interface{})
		go func() {
			defer close(outStream)
			for _, v := range values {
				time.Sleep(500 * time.Millisecond)

				select {
				case <-done:
					return
				case outStream <- v:
				}

			}
		}()
		return outStream
	}

	done := make(chan interface{})
	go func() {
		time.Sleep(10 * time.Second)
		close(done)
	}()
	nums := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9}
	result := doWork(done, generator(done, nums...))
	for res := range result {
		fmt.Println(res)
	}
	fmt.Println("exiting")
}
```
