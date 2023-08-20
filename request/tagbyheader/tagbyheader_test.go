package tagbyheader

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/dkotik/oakratelimiter/test"
)

func requestFactory(name, value string) test.RequestFactory {
	return func(ctx context.Context) *http.Request {
		r := test.GetRequestFactory(ctx)
		r.Header.Set(name, value)
		return r
	}
}

func TestRateLimitingByHeader(t *testing.T) {
	ctx := context.Background()
	l, err := New(
		WithName("test"),
		WithNewRate(5, time.Millisecond*50),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("try with first header", func(t *testing.T) {
		rf := requestFactory("test", "firstValue")
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

	t.Run("try with second header", func(t *testing.T) {
		rf := requestFactory("test", "secondValue")
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

	t.Run("try without a header", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			_, ok, err := l.Take(test.GetRequestFactory(ctx))
			if err != nil {
				t.Fatal("request limiter failed:", err)
			}
			if !ok {
				t.Fatal("request limiter blocked unexpectedly")
			}
		}
	})
}
