package tagbyip

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/dkotik/oakratelimiter/test"
)

func requestFactory(address string) test.RequestFactory {
	return func(ctx context.Context) *http.Request {
		r := test.GetRequestFactory(ctx)
		r.RemoteAddr = address
		return r
	}
}

func TestRateLimitingByIPAddress(t *testing.T) {
	ctx := context.Background()
	l, err := New(
		WithNewRate(5, time.Millisecond*50),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("try with first address", func(t *testing.T) {
		rf := requestFactory("127.0.0.1:8181")
		for i := 0; i < 5; i++ {
			_, ok, err := l.Take(rf(ctx))
			if err != nil {
				t.Fatal("request limiter failed:", err)
			}
			if !ok {
				t.Fatal("request limiter blocked unexpectedly")
			}
		}
	})

	t.Run("try with second address", func(t *testing.T) {
		rf := requestFactory("254.1.127.67:6775")
		for i := 0; i < 5; i++ {
			_, ok, err := l.Take(rf(ctx))
			if err != nil {
				t.Fatal("request limiter failed:", err)
			}
			if !ok {
				t.Fatal("request limiter blocked unexpectedly")
			}
		}
	})
}
