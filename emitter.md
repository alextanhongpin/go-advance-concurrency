```go
package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

const EventGreet = "greet"

type EmitterFunc func(in interface{}) error

type Emitter interface {
	Emit(evt string, in interface{})
	On(evt string, fn EmitterFunc)
	Start()
	Stop()
}

type Service interface {
	Emitter
	DoWork()
}

type ServiceImpl struct {
	*EmitterImpl
}

func (s *ServiceImpl) DoWork() {
	fmt.Println("do work")
	s.EmitterImpl.Emit(Event{EventGreet, "hallo!"})
}

func NewService() *ServiceImpl {
	return &ServiceImpl{EmitterImpl: NewEmitter(10000)}
}

type Event struct {
	Name    string
	Request interface{}
}

type EmitterImpl struct {
	ch     chan Event
	events map[string]EmitterFunc
	quit   chan interface{}
	sync.WaitGroup
	id int
	sync.RWMutex
}

func NewEmitter(n int) *EmitterImpl {
	return &EmitterImpl{
		ch:     make(chan Event, n),
		events: make(map[string]EmitterFunc),
		quit:   make(chan interface{}),
	}
}

func (e *EmitterImpl) work() error {
	select {
	case evt, ok := <-e.ch:
		if !ok {
			return errors.New("closed ch")

		}
		fn, exist := e.events[evt.Name]
		if !exist {
			return errors.New("not implemented")
		}
		return fn(evt.Request)
	}
}

func (e *EmitterImpl) Start() {
	e.Lock()

	n := e.id
	e.id++

	e.Add(1)
	e.Unlock()

	go func() {
		defer e.Done()
		for {
			select {
			case <-e.quit:
				return
			case evt, ok := <-e.ch:
				if !ok {
					log.Println(errors.New("closed ch"))
					return
				}
				fn, exist := e.events[evt.Name]
				if !exist {
					log.Println(errors.New("not implemented"))
				}
				if err := fn(evt.Request); err != nil {
					log.Println(err)
				} else {
					log.Println("worker", n, "processed")
				}
			}
		}
	}()
}

func (e *EmitterImpl) Stop() {
	// Should close the receiver
	// close(e.quit)
	fmt.Println("stopping service")
	close(e.ch)
	e.Wait()
}

func (e *EmitterImpl) Emit(evt Event) {
	select {
	case <-e.quit:
		return
	case e.ch <- evt:
		return
	case <-time.After(5 * time.Second):
		return
	}
}

func (e *EmitterImpl) On(evt string, fn EmitterFunc) {
	e.events[evt] = fn
}

func main() {
	svc := NewService()
	// Start the background worker for event emitter.
	svc.Start()
	// Can start multiple.
	svc.Start()
	svc.On(EventGreet, func(in interface{}) error {
		time.Sleep(1 * time.Second)
		switch v := in.(type) {
		case string:
			fmt.Println("got msg", v)
		}
		return nil
	})
	for i := 0; i < 10; i++ {
		svc.DoWork()
	}

	// Close the event emitter channel.
	svc.Stop()
	fmt.Println("terminating")
}
```
