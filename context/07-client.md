# Request With Context

## Status

`active`

## Context

Example of how to apply context cancelation when making http requests.


```go
package main

import (
 "context"
 "fmt"
 "math/rand"
 "net/http"
 "net/http/httptest"
 "time"
)

type result[T any] struct {
 data T
 err  error
}

func count(ctx context.Context, n int) <-chan result[int] {
 ch := make(chan result[int])
 go func() {
     defer close(ch)

     for i := 0; i < n; i++ {
         select {
         case <-ctx.Done():
             fmt.Println("terminating count ...")
             return
         default:
             ms := time.Duration(rand.Intn(50)) * time.Millisecond
             fmt.Println("sleep", ms, i)
             time.Sleep(ms)
             ch <- result[int]{data: i}
         }
     }
 }()

 return ch
}

func main() {
 defer func(start time.Time) {
     fmt.Println("took", time.Since(start))
 }(time.Now())
 ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
     ch := count(r.Context(), 5)

     var res []int
     for n := range ch {
         res = append(res, n.data)
     }

     fmt.Fprintf(w, "hello world: %v", res)
 }))
 defer ts.Close()

 ctx := context.Background()
 ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
 defer cancel()

 req, err := http.NewRequestWithContext(ctx, "GET", ts.URL, nil)
 if err != nil {
     panic(err)
 }

 client := http.Client{
     Timeout: 1 * time.Second,
 }
 resp, err := client.Do(req)
 if err != nil {
     fmt.Println("cause:", context.Cause(req.Context()))
     panic(err)
 }
 fmt.Println(resp.Status)
}
```
