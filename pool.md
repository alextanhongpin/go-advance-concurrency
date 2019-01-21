# Basic
```golang
package main

import (
	"fmt"
	"sync"
)

func main() {
	var i int
	pool := &sync.Pool{
		New: func() interface{} {
			i++
			fmt.Println("Creating new instance", i)
			return struct{}{}
		},
	}
	// Populate the pool with 4 instances first.
	pool.Put(pool.New())
	pool.Put(pool.New())
	pool.Put(pool.New())
	pool.Put(pool.New())

	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			instance := pool.Get()
			defer pool.Put(instance)
			// Do something quick here with the pool.
			// If the works takes longer to be returned and
			// another work is requesting item from the pool,
			// a new instance will be created instead,
			// hence defeating the purpose.
			// time.Sleep(1 * time.Second)
		}()
	}
	wg.Wait()
	fmt.Println("exiting")
}
```


## Example of pre-warming cache connection

```golang
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {

	var i int
	connToService := func() interface{} {
		i++
		fmt.Println("connecting...")
		time.Sleep(1 * time.Second)
		return struct{}{}
	}

	warmServiceConnCache := func(n int) *sync.Pool {
		c := &sync.Pool{New: connToService}
		for i := 0; i < n; i++ {
			c.Put(c.New())
		}
		return c
	}

	// Do work
	var wg sync.WaitGroup
	wg.Add(100)
	pool := warmServiceConnCache(3)
	for i := 0; i < 100; i++ {
		go func(i int) {
			defer wg.Done()
			instance := pool.Get()
			defer pool.Put(instance)
			fmt.Println("do work", i)
		}(i)
	}
	wg.Wait()
	fmt.Println(i, "instance created")
	fmt.Println("exiting")
}
```
