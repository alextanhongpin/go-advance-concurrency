```go
// You can edit this code!
// Click here and start typing.
package main

import (
	"fmt"
	"sync"
)

func main() {
	n := 10
	f := &Flush{}
	f.Add(n)

	s := make([]int, n)
	for i := range n {
		f.Go(func() {
			defer fmt.Println("done", i)
			fmt.Println("do", i)
			s[i] = i
			f.Notify()
			fmt.Println("got", s[i])
		})
	}
	f.Wait(func() {
		fmt.Println("got", s)
		for i, v := range s {
			s[i] = v * 10
		}
		fmt.Println("modify", s)
	})
}

type Flush struct {
	signal sync.WaitGroup
	wg     sync.WaitGroup
	waiter sync.WaitGroup
}

func (f *Flush) Add(n int) {
	f.signal.Add(n)
	f.wg.Add(n)
	f.waiter.Add(1)
}

func (f *Flush) Notify() {
	f.signal.Done()
	f.waiter.Wait()
}

func (f *Flush) Go(fn func()) {
	go func() {
		defer f.wg.Done()
		fn()
	}()
}

func (f *Flush) Wait(fn func()) {
	defer f.wg.Wait()
	f.signal.Wait()
	// do sth
	fn()
	f.waiter.Done()
}
```
