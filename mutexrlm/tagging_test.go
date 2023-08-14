package oakratelimiter

import (
	"context"
	"testing"
	"time"
)

func TestSingleTaggingMiddleware(t *testing.T) {
	limit := float64(2)
	interval := time.Millisecond * 20
	s, err := NewSingleTagging(
		WithSupervisingLimit(WithRate(limit, interval)),
		WithIPAddressTagger(WithRate(limit, interval)),
	)
	if err != nil {
		t.Fatal("unable to initialize single tagging rate limiter:", err)
	}
	mw := s.Middleware()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	t.Run("at a good rate", MiddlewareLoadTest(
		ctx,
		mw,
		NewRate(limit, interval+5),
		GetRequestFactory,
		0, // expected rejection rate
	))

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	t.Run("at a bad rate", MiddlewareLoadTest(
		ctx,
		mw,
		NewRate(limit*5, interval),
		GetRequestFactory,
		0.8, // expected rejection rate
	))
}

func TestMultiTaggingMiddleware(t *testing.T) {
	limit := float64(2)
	interval := time.Millisecond * 20
	m, err := NewMultiTagging(
		WithSupervisingLimit(WithRate(limit, interval)),
		WithIPAddressTagger(WithRate(limit, interval)),
		WithCookieTagger("test", "", WithRate(limit, interval)),
	)
	if err != nil {
		t.Fatal("unable to initialize single tagging rate limiter:", err)
	}
	mw := m.Middleware()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	t.Run("at a good rate", MiddlewareLoadTest(
		ctx,
		mw,
		NewRate(limit, interval+5),
		GetRequestFactory,
		0, // expected rejection rate
	))

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	t.Run("at a bad rate", MiddlewareLoadTest(
		ctx,
		mw,
		NewRate(limit*5, interval),
		GetRequestFactory,
		0.8, // expected rejection rate
	))
}
