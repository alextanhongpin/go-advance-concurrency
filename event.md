#  Workflow

Whenever we call bar's start method, we want to register bar to foo. And when we call bar's stop method, we want to deregister: 

```go
package main

import (
	"fmt"
)

type Foo struct {
}

func (f *Foo) register(bar *Bar) {
	fmt.Println("register")
}

func (f *Foo) deregister(bar *Bar) {
	fmt.Println("deregister")
}

func (f *Foo) Make() (*Bar, func()) {
	bar := new(Bar)
	f.register(bar)
	// What if we have multiple steps?

	return bar, func() {
		f.deregister(bar)
	}
}

type Bar struct {
}

func (b *Bar) Start() {
	fmt.Println("start")
}
func (b *Bar) Stop() {
	fmt.Println("stop")
}
func main() {
	fmt.Println("Hello, playground")
}
```

Using events channel to signal work:

```go
package main

import (
	"fmt"
	"log"
	"reflect"
	"sync"
)

type Foo struct {
	wg sync.WaitGroup
}

func (f *Foo) Close() {
	f.wg.Wait()
	fmt.Println("closed")
}

func (f *Foo) register(bar *Bar) {
	fmt.Println("register", bar)
}

func (f *Foo) deregister(bar *Bar) {
	fmt.Println("deregister", bar)
}

func (f *Foo) Make() *Bar {
	bar := NewBar()

	f.wg.Add(1)

	go func() {
		defer f.wg.Done()

		ch := bar.Events()
		for evt := range ch {
			switch e := evt.(type) {
			case BarStarted:
				fmt.Println("event", e.TypeName)
				f.register(bar)
			case BarStopped:
				fmt.Println("event", e.TypeName)
				f.deregister(bar)
				close(ch)
			default:
				log.Println("not implemented", e)
			}
		}
		fmt.Println("closing")
	}()
	
	return bar
}

type Event interface {
	GetTypeName() string
}

func getTypeName(i interface{}) string {
	if t := reflect.TypeOf(i); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

type BarStarted struct {
	TypeName string
}

func (b BarStarted) GetTypeName() string {
	return b.TypeName
}

func NewBarStarted() BarStarted {
	return BarStarted{
		TypeName: getTypeName(new(BarStarted)),
	}
}

type BarStopped struct {
	TypeName string
}

func NewBarStopped() BarStopped {
	return BarStopped{
		TypeName: getTypeName(new(BarStopped)),
	}
}

func (b BarStopped) GetTypeName() string {
	return b.TypeName
}

type Bar struct {
	evtCh chan Event
}
func NewBar() *Bar{
	return &Bar{
		evtCh: make(chan Event),
	}
}

func (b *Bar) Events() chan Event {
	return b.evtCh
}
func (b *Bar) Start() {
	fmt.Println("start")
	b.evtCh <- NewBarStarted()
}

func (b *Bar) Stop() {
	fmt.Println("stop")
	b.evtCh <- NewBarStopped()
}

func main() {
	f := new(Foo)
	defer f.Close()

	bar := f.Make()
	bar.Start()
	bar.Stop()
}
```
