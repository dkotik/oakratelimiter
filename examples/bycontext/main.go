//revive:disable:package-comments
package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/dkotik/oakratelimiter"
	"github.com/dkotik/oakratelimiter/request/tagbycontext"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

func main() {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	limiter, err := oakratelimiter.New(
		oakratelimiter.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) error {
				_, err := io.WriteString(w, "Hello World")
				return err
			},
		),
		oakratelimiter.WithContextTagger(
			tagbycontext.WithKey(contextKey),
			tagbycontext.WithNewRate(1, time.Second),
		),
	)
	if err != nil {
		panic(err)
	}

	server := &http.Server{
		Handler: limiter,
		BaseContext: func(_ net.Listener) context.Context {
			// normally use request.WithContext(...) inside handler
			//
			// for testing we
			// inject the same context value into all requests:
			return context.WithValue(
				context.Background(),
				contextKey,
				"testContextValue",
			)
		},
	}
	fmt.Printf("Listening at http://%s\n", l.Addr())
	server.Serve(l)
}
