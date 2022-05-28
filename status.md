# Status with atomic

```go
// You can edit this code!
// Click here and start typing.
package main

import (
	"fmt"
	"sync/atomic"
)

type Status int32

const (
	Pending Status = iota
	Rejected
	Fulfilled
)

func (s Status) Int32() int32 {
	return int32(s)
}

func (s Status) String() string {
	switch s {
	case Pending:
		return "pending"
	case Rejected:
		return "rejected"
	case Fulfilled:
		return "fulfilled"
	default:
		return "unknown"
	}
}

func (s Status) Valid() bool {
	return s >= Pending && s <= Fulfilled
}

type Task struct {
	status int32
}

func (t *Task) Status() Status {
	return Status(t.status)
}
func (t *Task) Resolve() bool {
	return atomic.CompareAndSwapInt32(&t.status, Pending.Int32(), Fulfilled.Int32())
}

func (t *Task) Reject() bool {
	return atomic.CompareAndSwapInt32(&t.status, Pending.Int32(), Rejected.Int32())
}

func main() {
	fmt.Println(Pending.Valid())
	fmt.Println(Rejected.Valid())
	fmt.Println(Fulfilled.Valid())
	fmt.Println(Status(-1).Valid())
	fmt.Println(Status(3).Valid())

	// pending -> resolved
	task := new(Task)
	fmt.Println(task.Resolve())
	fmt.Println(task.Status())
	fmt.Println(task.Reject())
	fmt.Println(task.Status())

	// pending -> rejected
	task = new(Task)
	task.Reject()
	fmt.Println(task.Status())
	fmt.Println(task.Resolve())
	fmt.Println(task.Status())
}
```

## Using sync.Once

Using `sync.Once`
