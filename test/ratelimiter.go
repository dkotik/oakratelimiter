package test

import (
	"testing"

	"github.com/dkotik/oakratelimiter/rate"
)

func RateLimiterTest(r rate.Limiter) func(t *testing.T) {
	return func(t *testing.T) {
		if r == nil {
			t.Fatalf("cannot use a %q rate limiter", r)
		}
		desiredRate := r.Rate()
		if desiredRate == rate.Zero {
			t.Fatalf("rate limiter %q has infinite desired rate", r)
		}

		t.Fatal("unimplemented")
	}
}
