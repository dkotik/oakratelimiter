package mutexrlm

import (
	"context"
	"testing"
	"time"

	"github.com/dkotik/oakratelimiter/test"
)

func TestRequestLimiter(t *testing.T) {
	limiter, err := NewRequestLimiter(WithNewRate(8, time.Millisecond*20))
	if err != nil {
		t.Fatal("cannot initialize request limiter:", err)
	}
	test.RequestLimiterTest(context.Background(), limiter, 8)(t)
}

// func TestBasicMiddleware(t *testing.T) {
// 	limit := float64(2)
// 	interval := time.Millisecond * 20
// 	basic, err := NewBasic(WithRate(limit, interval))
// 	if err != nil {
// 		t.Fatal("unable to initialize basic rate limiter:", err)
// 	}
// 	mw := basic.Middleware()
//
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
// 	defer cancel()
// 	t.Run("at a good rate", MiddlewareLoadTest(
// 		ctx,
// 		mw,
// 		NewRate(limit, interval+5),
// 		GetRequestFactory,
// 		0, // expected rejection rate
// 	))
//
// 	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()
// 	t.Run("at a bad rate", MiddlewareLoadTest(
// 		ctx,
// 		mw,
// 		NewRate(limit*5, interval),
// 		GetRequestFactory,
// 		0.8, // expected rejection rate
// 	))
// }
//
// func TestBasicRateLimiter(t *testing.T) {
// 	limit := float64(2)
// 	interval := time.Second
// 	rl, err := NewBasic(WithRate(limit, interval))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	cases := []struct {
// 		Sleep time.Duration
// 		Fails bool
// 	}{
// 		{Sleep: 0, Fails: false},
// 		{Sleep: 0, Fails: false},
// 		{Sleep: 0, Fails: true},
// 		{Sleep: 0, Fails: true},
// 		{Sleep: time.Millisecond * 500, Fails: false},
// 		{Sleep: time.Millisecond * 500, Fails: false},
// 		{Sleep: 0, Fails: true},
// 	}
//
// 	for i, c := range cases {
// 		time.Sleep(c.Sleep)
// 		err = rl.Take(nil)
// 		if err != nil && !c.Fails {
// 			t.Fatal(i+1, "rate limiter failed when not expecting it", rl.tokens, err)
// 		}
// 	}
// }
