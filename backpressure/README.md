# Backpressure

## Status

`proposed`


## Context

We want to apply backpressure, and drop requests when we can no longer handle incoming requests.


## Decision

We will implement backpressure as a middleware. The backpressure can be applied at selected paths after determining which endpoints are more prone to high load.


## Consequences


The system should be able to shed some load and handle request without failing.

However, requests will be dropped for some clients until the system is able to serve them.


Running the example will show the following:

```bash
➜  go-advance-concurrency git:(master) ✗ go run backpressure/main.go
send: 50
drop: 11
drop: 14
drop: 17
drop: 20
drop: 22
drop: 24
drop: 27
drop: 28
drop: 33
drop: 37
drop: 39
drop: 42
drop: 45
recv: 37 worker: 0
program exiting
```

Here, we send `50` requests. However, only `37` is received. The remainder has been dropped because the system cannot handle the request capacity.
