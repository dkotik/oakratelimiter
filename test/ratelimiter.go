package test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/dkotik/oakratelimiter/rate"
)

func RateLimiterTest(
	ctx context.Context,
	r rate.Limiter,
	parallel int,
) func(*testing.T) {
	return func(t *testing.T) {
		deadline, ok := ctx.Deadline()
		now := time.Now()
		if !ok {
			deadline = now.Add(time.Second * 5)
			var cancel func()
			ctx, cancel = context.WithDeadline(ctx, deadline)
			defer cancel()
			// t.Log()
		}
		halfTimeout := deadline.Sub(now) / 2
		if halfTimeout < r.Rate().Interval()+(time.Millisecond*200) {
			t.Fatal("not enough context time to run request limiter test:", halfTimeout.Seconds(), "seconds")
		}

		if r == nil {
			t.Fatal("cannot use a <nil> request limiter")
		}
		if parallel < 1 {
			t.Fatal("parallel tests setting must be greater than one")
		}

		halfContext, halfCancel := context.WithTimeout(ctx, halfTimeout)
		defer halfCancel()
		var wg sync.WaitGroup
		sleep := time.Duration(float64(r.Rate().Interval()) * 1.05)
		for i := 0; i < parallel; i++ {
			wg.Add(1)
			go func(ctx context.Context, r rate.Limiter) {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					default:
						remaining, ok, err := r.Take(ctx, "test", 1.0)
						if err != nil {
							t.Fatal(err)
							return
						}
						if !ok {
							t.Fatal("rate limiter maxed out, with long sleep:", remaining, "remaining")
							return
						}
						t.Logf("%.2f tokens remaining", remaining)
						time.Sleep(sleep)
					}
				}
			}(halfContext, r)
			time.Sleep(time.Millisecond * 5 * time.Duration(i+3))
		}
		wg.Wait()

		t.Log("testing request blocking")
		wg = sync.WaitGroup{}
		sleep = time.Duration(float64(r.Rate().Interval()) * 0.10)
		for i := 0; i < parallel; i++ {
			wg.Add(1)
			go func(ctx context.Context, r rate.Limiter) {
				blocked := 0
				passed := 0
				defer func() {
					wg.Done()
					if blocked == 0 {
						t.Fatal("request limiter never blocked even once")
					}
					percent := blocked * 100 / (passed + blocked)
					if percent < 55 {
						t.Fatalf(
							"requests were blocked %d%% of the time, unexpectedly",
							percent,
						)
					}
					t.Logf(
						"requests were blocked %d%% of the time, as expected",
						percent,
					)
				}()

				for {
					select {
					case <-ctx.Done():
						// t.Log("context done")
						return
					default:
						remaining, ok, err := r.Take(ctx, "test", 1.0)
						if err != nil {
							t.Fatal(err)
							return
						}
						if ok {
							passed++
						} else {
							blocked++
						}
						t.Logf("%.2f tokens remaining", remaining)
						time.Sleep(sleep)
					}
				}
			}(ctx, r)
			time.Sleep(time.Millisecond * 5 * time.Duration(i+3))
		}
		wg.Wait()
	}
}
