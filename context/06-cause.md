#  Context With Cancel Cause


## Status

`active`


## Context

Use context with cancel cause to return error when cancelling context manually. This allows it to be distinguish from the current error `context canceled.`.


## Decision

`ctx.Err()` should no longer be used. Instead, use `context.Cause(ctx)` to determine the cause of the error. We can still obtain the `context.Canceled` from `context.Cause` by calling `cancel(nil)`.


## Consequences

We can now tell if a context is canceled for a more specific reason.

```go
package main

import (
 "context"
 "errors"
 "fmt"
)

func main() {
 ctx := context.Background()

 {
     fmt.Println("Cancel with cause")
     ctx, cancel := context.WithCancelCause(ctx)
     cancel(errors.New("manual cancel"))
     debugContext(ctx)
 }

 {
     fmt.Println("Cancel with nil")
     ctx, cancel := context.WithCancelCause(ctx)
     cancel(nil)
     debugContext(ctx)
 }

 {
     fmt.Println("Cancel")
     ctx, cancel := context.WithCancel(ctx)
     cancel()
     debugContext(ctx)
 }
}

func debugContext(ctx context.Context) {
 fmt.Println("Error:", ctx.Err())
 fmt.Println("Cause:", context.Cause(ctx))
 fmt.Println()
}
```
