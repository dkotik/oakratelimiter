package mutexrlm

import (
	"context"
	"testing"
	"time"

	"github.com/dkotik/oakratelimiter/test"
)

func TestRateLimiter(t *testing.T) {
	limiter, err := New(WithNewRate(8, time.Millisecond*20))
	if err != nil {
		t.Fatal("cannot initialize request limiter:", err)
	}
	test.RateLimiterTest(context.Background(), limiter, 8)(t)
}
