## With switch

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type A struct{}
type B struct{}

func main() {
	bg := NewBackground()
	bg.Start()

	go func() {
		bg.Send(A{})
		bg.Send(&A{})
		bg.Send(B{})
		bg.Send(&B{})
	}()
	time.Sleep(1 * time.Second)
	bg.Stop()

	fmt.Println("terminating")
}

type Background struct {
	wg   *sync.WaitGroup
	ch   chan interface{}
	stop chan interface{}
}

func NewBackground() *Background {
	return &Background{
		wg:   new(sync.WaitGroup),
		ch:   make(chan interface{}),
		stop: make(chan interface{}),
	}

}

func (b *Background) Stop() {
	close(b.stop)
	b.wg.Wait()
}

func (b *Background) Start() {
	b.wg.Add(1)

	go func() {
		defer b.wg.Done()
		for {
			select {
			case <-b.stop:
				return
			case in, ok := <-b.ch:
				if !ok {
					return
				}
				switch v := in.(type) {
				case A:
					fmt.Println("got a", v)
				case *A:
					fmt.Println("got *a", v)
				case B:
					fmt.Println("got b", v)
				case *B:
					fmt.Println("got *b", v)
				default:
					fmt.Println("unhandled")
				}
			}
		}
	}()
}

func (b *Background) Send(in interface{}) {
	select {
	case <-b.stop:
		return
	case b.ch <- in:
		fmt.Println("receive work")
	case <-time.After(5 * time.Second):
		return
	}
}
```
## With Task

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type Task interface {
	Do() error
}
type TaskA struct {
	Count int
}

func (a *TaskA) Do() error {
	a.Count++
	fmt.Println(a.Count)
	return nil
}

func main() {
	bg := NewBackground()
	bg.Start()

	go func() {
		// Must be pointer receiver.
		// bg.Send(TaskA{})
		bg.Send(&TaskA{})
	}()
	time.Sleep(1 * time.Second)
	bg.Stop()

	fmt.Println("terminating")
}

type Background struct {
	wg   *sync.WaitGroup
	ch   chan Task
	stop chan interface{}
}

func NewBackground() *Background {
	return &Background{
		wg:   new(sync.WaitGroup),
		ch:   make(chan Task),
		stop: make(chan interface{}),
	}

}

func (b *Background) Stop() {
	close(b.stop)
	b.wg.Wait()
}

func (b *Background) Start() {
	b.wg.Add(1)

	go func() {
		defer b.wg.Done()
		for {
			select {
			case <-b.stop:
				return
			case in, ok := <-b.ch:
				if !ok {
					return
				}
				if err := in.Do(); err != nil {
					fmt.Println(err)
				}
			}
		}
	}()
}

func (b *Background) Send(in Task) {
	select {
	case <-b.stop:
		return
	case b.ch <- in:
		fmt.Println("receive work")
	case <-time.After(5 * time.Second):
		return
	}
}

```

## Communication through channel
```go
package main

import (
	"fmt"
)

type Request struct {
	Name string
}

type Task interface {
	Do(in interface{}) error
}

type (
	ServiceAOnly interface {
		A()
	}
	ServiceA struct {
		Count int
		svc   ServiceAOnly
	}
)

type (
	Service interface {
		A()
		B()
	}
	ServiceImpl struct{}
)

func (s *ServiceImpl) A() {
	fmt.Println("a")
}

func (s *ServiceImpl) B() {
	fmt.Println("b")
}

func (s *ServiceA) Do(in interface{}) error {
	s.Count++
	switch v := in.(type) {
	case Request:
		s.svc.A()
		fmt.Println(s.Count, v.Name)
	default:
		fmt.Println("unhandled")
	}
	return nil
}

var call = B("call")

func main() {
	bg := NewBackground()
	service := &ServiceImpl{}
	bg.Once(call, &ServiceA{Count: 100, svc: service})

	bg.Emit(call, Request{"request"})
	bg.Emit(call, Request{"s"})

	fmt.Println("terminating")
}

type B string
type Background struct {
	events map[B]Task
}

func NewBackground() *Background {
	return &Background{
		events: make(map[B]Task),
	}
}

func (b *Background) Once(evt B, fn Task) {
	b.events[evt] = fn
}

func (b *Background) Emit(evt B, payload interface{}) {
	fn, exist := b.events[evt]
	if !exist {
		return
	}

	err := fn.Do(payload)
	if err != nil {
		fmt.Println(err)
	}
}
```
