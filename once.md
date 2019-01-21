```golang
package main

import (
	"fmt"
	"sync"
)

const N = 100

func main() {
	var i int
	increment := func() {
		i += 1
	}

	var once sync.Once
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			once.Do(increment)
		}()
	}
	wg.Wait()
	fmt.Println("result is:", i)
}
```

## Singleton pattern with sync.Once

Advantages over `init`:

- you can perform lazy-initialization of variables
- you can pass in parameters through runtime

```go
package main

import (
	"fmt"
	"sync"
)
type singleton struct{}

var instance *singleton

var once sync.Once

func GetInstance() *singleton {
	once.Do(func() {
		instance = &singleton{}
	})
	return instance
}

func main() {
	sgt := GetInstance()
	fmt.Println(sgt)
}
```
