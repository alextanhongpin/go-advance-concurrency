# mastering-go-concurrency
Go Concurrency 101


More in the [wiki](https://github.com/alextanhongpin/go-advanced-concurrency/wiki).

- Concurrency is the independent executions of goroutines.
- Functions are created as goroutines with the keyword go.
- Goroutines are executed within the scope of a logical processor that owns a single operating system thread and run queue.
- A race conditions is when two or more goroutines attempt to access the same resource.
- Atomic functions and mutexes provide a safe way to protect against race conditions.
- Channels provide an intrinsic way to safely share data between two goroutines.
- Unbuffered channels provide a guarantee between an exchange of data. Buffered channels do not.


# Book: Concurrency in Go


## Deadlocks
  A deadlock program is one in which all concurrent processes are waiting on one another. In this state, the program will never recover without outside intervention.

There are a few conditions for deadlock to occur, which are called Coffman Conditions.

1. Mutual exclusions: A concurrent process holds exclusive rights to a resource at any one time
2. Wait for Condition: A concurrent process must simultaneously hold a resource and be waiting for an additional resource
3. No preemption: A resource held by a concurrent process can only be released by that process, so it fulfills the conditions
4. Circular wait: A concurrent process (P1) must be waiting on a chain of other concurrent processes (P2), which are in turn waiting on it (P1), so it fulfills this final condition too.

## Livelocks

  Livelocks are programs that are actively performing concurrent operations, but these operations do nothing to move the state of the program forward.


## Starvation

Starvation is any situation where a concurrent process cannot get all the resources it needs to perform work.

## Data Race Issue

### Returning slice getters

This will cause data race, since slice is a reference even when it is passed as value:

```go
func (b *EpsilonGreedy) GetCounts() []int {
 	b.RLock()
 	defer b.RUnlock()

	return b.Counts
}
```

Correct way:

```go
func (b *EpsilonGreedy) GetCounts() []int {
 	b.RLock()
 	defer b.RUnlock()

	sCopy := make([]int, len(b.Counts))
	copy(sCopy, b.Counts)
	return sCopy
}
```

## TODO

- gracefully shutting down channels
- ensure all the consumer channels are flushed before shutting them down
- how to register different channel type for events
- how to create event emitter like nodejs


## Concurrency Patterns

- producer-consumer
- active object
- monitor object
- half-sync/half-async
- leader/followers
- balking pattern
- barrier
- double-checked locking
- guarded suspension
- nuclear reaction
- reactor pattern
- read write lock pattern
- scheduler pattern
- thread pool pattern
- thread-local storage

## References

- https://en.wikipedia.org/wiki/Concurrency_pattern
- https://sudo.ch/unizh/concurrencypatterns/ConcurrencyPatterns.pdf
