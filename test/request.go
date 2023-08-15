package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
)

// RequestFactory generates new requests for load testing rate limiting middleware. Use together with [MiddlewareLoadTest].
type RequestFactory func(context.Context) *http.Request

// GetRequestFactory is the simplest request factory with no payload.
func GetRequestFactory(ctx context.Context) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	return r.WithContext(ctx)
}

func RequestFanOut(
	ctx context.Context,
	in chan *http.Request,
	out []chan *http.Request,
) error {
	for {
		for _, c := range out {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case request := <-in:
				c <- request
				// continue
			}
		}
	}
	return nil
}

func ErrorsFanIn(ctx context.Context, in []chan error) (out chan error) {
	var wg sync.WaitGroup
	wg.Add(len(in))
	out = make(chan error)
	for _, c := range in {
		go func(c <-chan error) {
			defer wg.Done()
			for err := range c {
				select {
				case <-ctx.Done():
					return
				case out <- err:
				}
			}
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
