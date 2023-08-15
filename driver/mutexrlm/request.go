package mutexrlm

import (
	"net/http"
	"sync"
	"time"

	"github.com/dkotik/oakratelimiter/rate"
)

// Basic rate limiter enforces the limit using one leaky token bucket.
type Basic struct {
	failure  error
	interval time.Duration
	rate     rate.Rate
	limit    float64

	mu sync.Mutex
	bucket
}

// Rate returns the rate limiter [Rate].
func (b *Basic) Rate() Rate {
	return b.rate
}

// Take consumes one token per request.
func (b *Basic) Take(r *http.Request) (
	remaining float64,
	ok bool,
	err error,
) {
	t := time.Now()
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.bucket.Take(b.limit, b.rate, t, t.Add(b.interval))
}
