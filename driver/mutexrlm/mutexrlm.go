/*
Package mutexrlm provides [rate.Limiter]s that use memory stores with [sync.Mutex] for safe concurrency. This strategy is optimal for simple single-instance rate limiting. Use multiple [RateLimiter]s on endpoints to avoid lock contention when dealing with large traffic volume.
*/
package mutexrlm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dkotik/oakratelimiter/rate"
)

// New initializes a [RateLimiter] using a list of [Option]s.
func New(withOptions ...Option) (*RateLimiter, error) {
	o := &options{}
	for _, option := range append(
		withOptions,
		WithDefaultBurst(),
		WithDefaultInitialAllocationSize(),
		WithDefaultCleanupInterval(),
		WithDefaultCleanupContext(),
	) {
		if err := option(o); err != nil {
			return nil, fmt.Errorf("cannot initialize mutex rate limiter driver: %w", err)
		}
	}

	r := &RateLimiter{
		rate:       o.Rate,
		burstLimit: o.Burst,
		mu:         sync.Mutex{},
		buckets: make(
			map[string]*rate.LeakyBucket,
			o.InitialAllocationSize,
		),
	}

	go func(ctx context.Context, every time.Duration, r *RateLimiter) {
		ticker := time.NewTicker(every)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				r.Purge(t)
			}
		}
	}(o.CleanupContext, o.CleanupInterval, r)

	return r, nil
}

type RateLimiter struct {
	rate       *rate.Rate
	burstLimit float64

	mu      sync.Mutex
	buckets map[string]*rate.LeakyBucket
}

func (r *RateLimiter) Rate() *rate.Rate {
	return r.rate
}

// Remaining locates the proper [rate.LeakyBucket] by tag returns the number of tokens still in it. If the bucket does not exist, returns the burst limit.
func (r *RateLimiter) Remaining(
	ctx context.Context,
	tag string,
) (
	remaining float64,
	err error,
) {
	r.mu.Lock()
	defer r.mu.Unlock()

	foundBucket, ok := r.buckets[tag]
	if !ok {
		return r.burstLimit, nil
	}
	foundBucket.Refill(time.Now(), r.rate, r.burstLimit)
	return foundBucket.Remaining(), nil
}

// Take locates the proper [rate.LeakyBucket] by tag and takes one token from it. If the bucket does not exist, a new one is added to the internal map.
func (r *RateLimiter) Take(
	ctx context.Context,
	tag string,
	tokens float64,
) (
	remaining float64,
	ok bool,
	err error,
) {
	t := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	foundBucket, ok := r.buckets[tag]
	if !ok {
		foundBucket = rate.NewLeakyBucket(
			t,
			r.rate,
			r.burstLimit,
		)
		r.buckets[tag] = foundBucket
	} else {
		foundBucket.Refill(t, r.rate, r.burstLimit)
	}
	remaining, ok = foundBucket.Take(tokens)
	return
}

// Purge removes all tokens that are expired by given [time.Time].
func (r *RateLimiter) Purge(at time.Time) {
	at = at.Add(-r.rate.Interval())
	r.mu.Lock()
	defer r.mu.Unlock()

	for k, bucket := range r.buckets {
		if bucket.Touched().Before(at) {
			delete(r.buckets, k)
		}
	}
}
