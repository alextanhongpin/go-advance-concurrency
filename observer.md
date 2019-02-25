## Observer Pattern

Background observer pattern implementation.

```go
package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type ObserverEvent string
type ObserverFunc func(interface{}) error
type Observer interface {
	On(event ObserverEvent, fn ObserverFunc)
	Emit(event ObserverEvent, params interface{}) bool
	Start()
	Stop()
}

type ObserverMessage struct {
	Name   ObserverEvent
	Params interface{}
}

type ObserverImpl struct {
	sync.WaitGroup
	events map[ObserverEvent][]ObserverFunc

	quit chan interface{}
	ch   chan ObserverMessage
}

func (o *ObserverImpl) On(event ObserverEvent, fn ObserverFunc) {
	_, exist := o.events[event]
	if !exist {
		o.events[event] = make([]ObserverFunc, 0)
	}
	o.events[event] = append(o.events[event], fn)
}

func (o *ObserverImpl) Emit(event ObserverEvent, params interface{}) bool {
	select {
	case <-o.quit:
		return false
	case o.ch <- ObserverMessage{event, params}:
		fmt.Println("snet")
		return true
	case <-time.After(5 * time.Second):
		return false
	}
}

func (o *ObserverImpl) Stop() {
	close(o.quit)
	o.Wait()
}

func (o *ObserverImpl) Start() {
	o.Add(1)
	go func() {
		defer o.Done()
		for {
			select {
			case <-o.quit:
				return
			case evt, ok := <-o.ch:
				if !ok {
					return
				}
				fns, exist := o.events[evt.Name]
				if !exist {
					log.Println(fmt.Errorf(`event "%s" does not exist`, evt.Name))
				}
				for _, fn := range fns {
					time.Sleep(1 * time.Second)
					if err := fn(evt.Params); err != nil {
						log.Println(err)
					}
				}
			}
		}
	}()
}

func NewObserver(n int) *ObserverImpl {
	events := make(map[ObserverEvent][]ObserverFunc)
	return &ObserverImpl{
		events: events,
		quit:   make(chan interface{}),
		ch:     make(chan ObserverMessage, n),
	}
}

const (
	GreetEvent  = ObserverEvent("greet")
	ScreamEvent = ObserverEvent("scream")
)

type UserService struct {
	Observer
}

func (u *UserService) Greet(msg string) {
	fmt.Println("greeting", msg)
	u.Emit(GreetEvent, msg)
}

type ScreamRequest struct {
	Msg string
}

func (u *UserService) Scream(req ScreamRequest) {
	// Do some work...
	u.Emit(ScreamEvent, req)
}

func main() {
	usersvc := &UserService{NewObserver(10)}
	usersvc.Start()

	usersvc.On(GreetEvent, func(msg interface{}) error {
		fmt.Println("got", msg)
		return nil
	})
	usersvc.Greet("hello!")

	usersvc.On(ScreamEvent, func(msg interface{}) error {
		switch v := msg.(type) {
		case ScreamRequest:
			fmt.Println("got scream msg:", v.Msg)
		default:
			fmt.Println("not handled")
		}
		return nil
	})

	go func() {
		for i := 0; i < 10; i++ {
			usersvc.Scream(ScreamRequest{"screammmm"})
		}
	}()

	time.Sleep(10 * time.Second)
	usersvc.Stop()
}
```
