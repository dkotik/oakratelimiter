# Oak Rate Limiter

<p>
	<a href="https://pkg.go.dev/github.com/dkotik/oakratelimiter"><img src="https://img.shields.io/badge/godoc-reference-%23007d9c.svg"></a>
</p>

Flexible HTTP rate limiter with multiple backend drivers and optional timing modulation with partial obfuscation.

## Usage

```go
package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/dkotik/oakratelimiter"
	"github.com/dkotik/oakratelimiter/driver/mutexrlm"
)

// see more examples in the examples directory
func main() {
  // select random port
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	rl, err := mutexrlm.NewRequestLimiter(mutexrlm.WithNewRate(1, time.Second))
	if err != nil {
		panic(err)
	}

	limiter, err := oakratelimiter.New(
		oakratelimiter.HandlerFunc( // Oak Handler
			func(w http.ResponseWriter, r *http.Request) error {
				_, err := io.WriteString(w, "Hello World")
				return err
			},
		),
		oakratelimiter.WithRequestLimiter("global", rl),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Listening at http://%s\n", l.Addr())
	http.Serve(l, limiter)
}
```

## Supported Backend Drivers

- [x] In-memory sync.Mutex map: `mutexrlmrlm.New`
- [x] Postgres: `postgresrlm.New`
- [x] SQLite: `sqliterlm.New`
- [ ] (planned) Swiss map
- [ ] Atomic
- [ ] Redis

## Bundled Request Taggers

Taggers differentiate requests based on a property. Each can be combined with a different backend driver.

- [x] By IP address: `tagbyip.New`
- [x] By Header: `tagbyheader.New`
  - [x] Supports optional `WithNoHeaderLimiter`
  - [ ] Detection of similar headers from common header set:
      - X-Originating-IP
      - X-Forwarded-For
      - X-Forwarded-Host
      - X-Remote-IP
      - X-Remote-Addr
      - X-Client-IP
      - X-Host
- [x] By Cookie: `tagbycookie.New`
  - [x] Supports optional `WithNoCookieLimiter`
- [x] By Context Value: `tagbycontext.New`
  - [x] Supports optional `WithNoValueLimiter`
- [ ] Ensured protection against double HTTP headers:
  ```
  X-Forwarded-For:
  X-Forwarded-For: 127.0.0.1
  ```
- [ ] Ensured protection against null bytes at the end of URL or headers:
  - [ ] %00
  - [ ] %0d
  - [ ] %0a
  - [ ] %0d%0a
  - [ ] %09
  - [ ] %0c
  - [ ] %20

## Timing Modulation

Use `timing.NewTimingModulator` or `timing.NewMiddleware` to protect endpoints from timing attacks by injecting random delays.
