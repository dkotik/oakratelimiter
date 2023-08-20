package tagbycontext

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/dkotik/oakratelimiter/test"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

func requestFactory(value string) test.RequestFactory {
	return func(ctx context.Context) *http.Request {
		withValue := context.WithValue(ctx, contextKey, value)
		r := test.GetRequestFactory(ctx).WithContext(withValue)
		return r
	}
}

func TestRateLimitingByContextValue(t *testing.T) {
	ctx := context.Background()
	l, err := New(
		WithKey(contextKey),
		WithNewRate(5, time.Millisecond*50),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("try with first context value", func(t *testing.T) {
		rf := requestFactory("firstValue")
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

	t.Run("try with second context value", func(t *testing.T) {
		rf := requestFactory("secondValue")
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

	t.Run("try with no context value", func(t *testing.T) {
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
