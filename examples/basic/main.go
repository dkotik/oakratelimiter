//revive:disable:package-comments
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

func main() {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	rl, err := mutexrlm.NewRequestLimiter(
		mutexrlm.WithNewRate(1, time.Second), // once per second
	)
	if err != nil {
		panic(err)
	}

	limiter, err := oakratelimiter.New(
		oakratelimiter.HandlerFunc(
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
