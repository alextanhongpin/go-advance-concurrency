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


## Generic Event Emitter

```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	work := func(ctx context.Context, evt SomeEvent) {
		fmt.Println("received:", evt)
		time.Sleep(100 * time.Millisecond)
	}

	s := newSynchronousEventHandler[SomeEvent]()
	s.Register(work)

	n := 10
	benchmark("sync took:", n, s, SomeEvent{})

	as := newAsynchronousEventHandler[SomeEvent](2) // 2x the speed
	defer as.Close()

	as.Register(work)
	benchmark("async took:", n, as, SomeEvent{})
}

func benchmark(msg string, n int, e EventEmitter[SomeEvent], evt SomeEvent) {
	defer func(start time.Time) {
		fmt.Println(msg, time.Since(start))
	}(time.Now())

	for i := 0; i < n; i++ {
		e.Emit(context.Background(), evt)
	}
}

// Instead of creating one event emitter for all events, we create one event
// emitter per event.
type EventEmitter[T any] interface {
	// Emit should not return any error.
	Emit(ctx context.Context, t T)
}

// EventHandler should not have error
type EventHandler[T any] func(ctx context.Context, t T)

type EventRegistrator[T any] interface {
	Register(handlers ...EventHandler[T])
}

type SomeEvent struct {
}

var _ interface {
	EventEmitter[SomeEvent]
	EventRegistrator[SomeEvent]
} = (*synchronousEventHandler[SomeEvent])(nil)

type synchronousEventHandler[T any] struct {
	handlers []EventHandler[T]
}

func newSynchronousEventHandler[T any]() *synchronousEventHandler[T] {
	return &synchronousEventHandler[T]{}
}

func (h *synchronousEventHandler[T]) Emit(ctx context.Context, evt T) {
	for i := 0; i < len(h.handlers); i++ {
		h.handlers[i](ctx, evt)
	}
}

func (h *synchronousEventHandler[T]) Register(handlers ...EventHandler[T]) {
	h.handlers = append(h.handlers, handlers...)
}

var _ interface {
	EventEmitter[SomeEvent]
	EventRegistrator[SomeEvent]
} = (*asynchronousEventHandler[SomeEvent])(nil)

type asynchronousEventHandler[T any] struct {
	handlers  []EventHandler[T]
	start     sync.Once
	stop      sync.Once
	ch        chan T
	done      chan bool
	wg        sync.WaitGroup
	numWorker int
}

func newAsynchronousEventHandler[T any](n int) *asynchronousEventHandler[T] {
	return &asynchronousEventHandler[T]{
		done:      make(chan bool),
		ch:        make(chan T),
		numWorker: n,
	}
}

func (h *asynchronousEventHandler[T]) Close() {
	h.stop.Do(func() {
		close(h.done)
		h.wg.Wait()
	})
	fmt.Println("closed")
}

func (h *asynchronousEventHandler[T]) init(n int) {
	if n == 0 {
		panic("at least one worker is required")
	}
	h.wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer h.wg.Done()

			for {
				select {
				case evt := <-h.ch:
					for i := 0; i < len(h.handlers); i++ {
						h.handlers[i](context.Background(), evt)
					}
				case <-h.done:
					return
				}
			}
		}()
	}
}

func (h *asynchronousEventHandler[T]) Emit(ctx context.Context, evt T) {
	// Skip if no handlers.
	if len(h.handlers) == 0 {
		return
	}

	h.start.Do(func() {
		h.init(h.numWorker)
	})

	h.ch <- evt
}

func (h *asynchronousEventHandler[T]) Register(handlers ...EventHandler[T]) {
	h.handlers = append(h.handlers, handlers...)
}
```
