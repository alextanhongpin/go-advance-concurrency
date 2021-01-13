## Broker pattern

Use this if you want to broadcast the same message to all subscribers. Alternatively, we can also register the subscriber by id. Then we can publish to the individual subscribers.

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type Broker struct {
	wg      sync.WaitGroup
	once    sync.Once
	done    chan struct{}
	pubCh   chan interface{}
	subCh   chan chan interface{}
	unsubCh chan chan interface{}
}

func NewBroker() *Broker {
	return &Broker{
		done:    make(chan struct{}),
		pubCh:   make(chan interface{}, 1),
		subCh:   make(chan chan interface{}, 1),
		unsubCh: make(chan chan interface{}, 1),
	}
}

func (b *Broker) Start() {
	subs := map[chan interface{}]struct{}{}
	b.wg.Add(1)

	go func() {
		defer b.wg.Done()
		for {
			select {
			case <-b.done:
				for msgCh := range subs {
					close(msgCh)
				}
				return
			case msgCh := <-b.subCh:
				subs[msgCh] = struct{}{}
			case msgCh := <-b.unsubCh:
				delete(subs, msgCh)
			case msg := <-b.pubCh:
				for msgCh := range subs {
					// msgCh is buffered, use non-blocking send to protect the broker.
					select {
					case msgCh <- msg:
					default:
					}
				}
			}
		}
	}()
}

func (b *Broker) Stop() {
	b.once.Do(func() {
		close(b.done)
		b.wg.Wait()
	})
}

func (b *Broker) Subscribe() chan interface{} {
	msgCh := make(chan interface{}, 5)
	b.subCh <- msgCh
	return msgCh
}

func (b *Broker) Unsubscribe(msgCh chan interface{}) {
	b.unsubCh <- msgCh
	close(msgCh)
}

func (b *Broker) Publish(msg interface{}) {
	b.pubCh <- msg
}

func main() {
	b := NewBroker()
	b.Start()

	// Create and subscribe 3 client func.
	clientFunc := func(id int) {
		// NOTE: We can also subscribe to a particular id.
		msgCh := b.Subscribe()
		for msg := range msgCh {
			fmt.Println(msg)
		}
	}
	for i := 0; i < 3; i++ {
		go clientFunc(i)
	}

	// Start publishing messages.
	go func() {
		for i := 0; i < 10; i++ {
			b.Publish(fmt.Sprintf("msg#%d", i))
			time.Sleep(300 * time.Millisecond)
		}
	}()

	select {
	case <-time.After(3 * time.Second):
		b.Stop()
	}

	fmt.Println("Hello, playground")
}
```

## Broadcast vs Publish

Create individual subscribers


```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type Subscriber struct {
	id string
	ch chan interface{}
}

type Message struct {
	id  string
	msg interface{}
}

type Broker struct {
	wg          sync.WaitGroup
	once        sync.Once
	done        chan struct{}
	pubCh       chan Message
	broadcastCh chan interface{}
	subCh       chan Subscriber
	unsubCh     chan string
}

func NewBroker() *Broker {
	return &Broker{
		done:        make(chan struct{}),
		pubCh:       make(chan Message, 1),
		broadcastCh: make(chan interface{}, 1),
		subCh:       make(chan Subscriber, 1),
		unsubCh:     make(chan string, 1),
	}
}

func (b *Broker) Start() {
	subs := map[string]chan interface{}{}
	b.wg.Add(1)

	go func() {
		defer b.wg.Done()
		for {
			select {
			case <-b.done:
				for _, msgCh := range subs {
					close(msgCh)
				}
				return
			case sub := <-b.subCh:
				subs[sub.id] = sub.ch
			case id := <-b.unsubCh:
				close(subs[id])
				delete(subs, id)
			case m := <-b.pubCh:
				subs[m.id] <- m.msg
			case msg := <-b.broadcastCh:
				for _, msgCh := range subs {
					// msgCh is buffered, use non-blocking send to protect the broker.
					select {
					case msgCh <- msg:
					default:
					}
				}
			}
		}
	}()
}

func (b *Broker) Stop() {
	b.once.Do(func() {
		close(b.done)
		b.wg.Wait()
	})
}

func (b *Broker) Subscribe(id string) chan interface{} {
	sub := Subscriber{
		id: id,
		ch: make(chan interface{}, 5),
	}
	b.subCh <- sub
	return sub.ch
}

func (b *Broker) Unsubscribe(id string) {
	b.unsubCh <- id
}

func (b *Broker) Publish(id string, msg interface{}) {
	b.pubCh <- Message{id: id, msg: msg}
}

func (b *Broker) Broadcast(msg interface{}) {
	b.broadcastCh <- msg
}

func main() {
	b := NewBroker()
	b.Start()

	// Create and subscribe 3 client func.
	clientFunc := func(id int) {
		// NOTE: We can also subscribe to a particular id.
		msgCh := b.Subscribe(fmt.Sprint(id))
		for msg := range msgCh {
			fmt.Println(msg)
		}
	}
	for i := 0; i < 3; i++ {
		go clientFunc(i)
	}

	// Publishing to specific subscribers.
	go func() {
		for i := 0; i < 10; i++ {
			b.Publish(fmt.Sprint(i%3), fmt.Sprintf("[subscriber-%d] msg#%d", i%3, i))
			time.Sleep(300 * time.Millisecond)
		}
	}()

	// Broadcast to all.
	go func() {
		for i := 0; i < 10; i++ {
			b.Broadcast(fmt.Sprintf("[broadcast] msg#%d", i))
			time.Sleep(300 * time.Millisecond)
		}
	}()

	select {
	case <-time.After(3 * time.Second):
		b.Stop()
	}

	fmt.Println("Hello, playground")
}
```
