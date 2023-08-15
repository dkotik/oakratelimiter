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
		if err := desiredRate.Validate(); err != nil {
			t.Fatal("invalid desired rate:", err)
		}

		t.Fatal("unimplemented")
	}
}
