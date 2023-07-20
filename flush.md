```go
package main

import "fmt"

func main() {
	ch := make(chan int, 5)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)

  // Always remember to flush the buffered channels after closing them.
	for len(ch) > 0 {
		fmt.Println(<-ch)
	}
}

```
