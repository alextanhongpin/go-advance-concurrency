# Periodic task with ticker

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()
	done := make(chan interface{})
	go func() {
		time.Sleep(5 * time.Second)
		close(done)
	}()
	for {
		select {
		case <-done:
			return
		case <-t.C:
			fmt.Println("do work")
		}
	}
}
```
