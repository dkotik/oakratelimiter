//revive:disable:package-comments
package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/dkotik/oakratelimiter"
	"github.com/dkotik/oakratelimiter/request/tagbycookie"
)

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
		oakratelimiter.WithCookieTagger(
			tagbycookie.WithName("sessionUUID"),
			tagbycookie.WithNewRate(1, time.Second),
		),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Listening at http://%s\n", l.Addr())
	fmt.Println("To test, run the following command in terminal:")
	fmt.Printf("curl -v --cookie \"sessionUUID=one\" http://%s\n", l.Addr())
	http.Serve(l, limiter)
}
