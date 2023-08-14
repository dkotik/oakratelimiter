package oakratelimiter

import (
	"context"
	"math"
	"time"
)

// Rate is the number of tokens leaked or replenished per nanosecond.
type Rate float64

// NewRate creates a [Rate] based on expected limit and a given time interval.
func NewRate(limit float64, interval time.Duration) Rate {
	if interval == 0 {
		return Rate(math.Inf(1))
	}
	return Rate(limit / float64(interval.Nanoseconds()))
}

// ReplenishedTokens returns fractional token amount based on time passed.
func (r Rate) ReplenishedTokens(from, to time.Time) float64 {
	return float64(to.Sub(from).Nanoseconds()) * float64(r)
}

// RateLimiter contrains the number of requests to a certain [Rate]. When it is exceeded, it should return [ErrTooManyRequests].
type RateLimiter interface {
	Rate() Rate
	Take(
		ctx context.Context,
		tag string,
		tokens float64,
	) (
		remaining float64,
		ok bool,
		err error,
	)
}
