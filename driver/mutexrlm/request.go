package mutexrlm

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/dkotik/oakratelimiter/rate"
	"github.com/dkotik/oakratelimiter/request"
)

// NewRequestLimiter initializes a [request.Limiter] using [Option]s.
func NewRequestLimiter(withOptions ...Option) (request.Limiter, error) {
	o := &options{}
	for _, option := range append(
		withOptions,
		WithDefaultBurst(),
		func(o *options) error { // validate
			if o.InitialAllocationSize != 0 {
				return errors.New("initial allocation option does not apply to a request limiter")
			}
			if o.CleanupContext != nil {
				return errors.New("clean up context option does not apply to a request limiter")
			}
			if o.CleanupInterval != 0 {
				return errors.New("clean up interval option does not apply to a request limiter")
			}
			return nil
		},
	) {
		if err := option(o); err != nil {
			return nil, err
		}
	}
	return &requestLimiter{
		rate:       o.Rate,
		burstLimit: o.Burst,
		mu:         sync.Mutex{},
		bucket:     *rate.NewLeakyBucket(time.Now(), o.Rate, o.Burst),
	}, nil
}

type requestLimiter struct {
	rate       *rate.Rate
	burstLimit float64

	mu     sync.Mutex
	bucket rate.LeakyBucket
}

// Rate returns the rate limiter [rate.Rate].
func (l *requestLimiter) Rate() *rate.Rate {
	return l.rate
}

// Take consumes one token per request.
func (l *requestLimiter) Take(r *http.Request) (
	remaining float64,
	ok bool,
	err error,
) {
	t := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.bucket.Refill(t, l.rate, l.burstLimit)
	remaining, ok = l.bucket.Take(1.0)
	return
}
