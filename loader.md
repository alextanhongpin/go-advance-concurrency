# Thread-safe Lazy Loading


```go
package main

import (
	"embed"
	"fmt"
	"path/filepath"
	"sync"
)

//go:embed data/*.txt
var data embed.FS

func main() {
	one, err := data.ReadFile("data/one.txt")
	if err != nil {
		panic(err)
	}

	two, err := data.ReadFile("data/two.txt")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(one))
	fmt.Println(string(two))
	l := &Loader{data: make(map[string][]byte)}
	fmt.Println(string(l.Load("one")))
	fmt.Println(string(l.Load("one")))
}

type Loader struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func (l *Loader) Load(name string) []byte {
	name = filepath.Join("data", name+".txt")

	l.mu.RLock()
	b, ok := l.data[name]
	l.mu.RUnlock()
	if ok {
		return b
	}

	// Run the heavy function.
	// This is done outside the lock to prevent holding the lock.
	b, err := data.ReadFile(name)
	if err != nil {
		panic(err)
	}

	// Start a lock to store the data.
	l.mu.Lock()
	// We also check if the data already exists before overwriting.
	if b, ok := l.data[name]; ok {
		l.mu.Unlock()
		return b
	}

	l.data[name] = b
	l.mu.Unlock()

	return b
}
```
