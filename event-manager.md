# Event Manager

Recording events in the background and persisting them using golang.

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type (
	E     map[string]interface{}
	Event struct {
		ID   string
		Name string
		Data E
	}

	EventManager struct {
		ch   chan Event
		stop chan struct{}
		wg   *sync.WaitGroup
	}
)

func main() {
	mgr := NewEventManager(10)
	go func() {
		t := time.NewTicker(250 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				evt := NewEvent("1", "hello")
				evt.Data = E{"metadata": "something"}
				mgr.Send(evt)
			}
		}
	}()
	mgr.Start()
	// Can start multiple.
	mgr.Start()
	time.Sleep(10 * time.Second)
	mgr.Stop()
}

func NewEvent(id, name string) Event {
	return Event{id, name, make(map[string]interface{})}
}

func NewEventManager(buffer int) *EventManager {
	return &EventManager{
		ch:   make(chan Event, buffer),
		stop: make(chan struct{}),
		wg:   new(sync.WaitGroup),
	}
}

func (e *EventManager) Start() {
	e.wg.Add(1)
	fmt.Println("starting event manager")
	go e.loop()
}

func (e *EventManager) Stop() {
	close(e.stop)
	e.wg.Wait()
	fmt.Println("quit")
}

func (e *EventManager) Send(evt Event) bool {
	select {
	case <-e.stop:
		return false
	case e.ch <- evt:
		return true
	case <-time.After(5 * time.Second):
		return false
	}
}

func (e *EventManager) loop() {
	defer e.wg.Done()
	for {
		select {
		case <-e.stop:
			fmt.Println("stopped")
			return
		case evt, ok := <-e.ch:
			fmt.Println(evt, ok)
		}
	}
}
```
