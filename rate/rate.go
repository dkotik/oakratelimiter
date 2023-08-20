/*
Package rate defines [Rate] and provides basic interfaces for rate limiter driver implementations.
*/
package rate

import (
	"errors"
	"fmt"
	"math"
	"time"

	"log/slog"
)

type Rate struct {
	tokens              float64
	tokensPerNanosecond float64
	interval            time.Duration
}

// New creates a [Rate] based on expected limit and a given time interval.
func New(limit float64, interval time.Duration) (*Rate, error) {
	rate := &Rate{
		tokens:              limit,
		tokensPerNanosecond: limit / float64(interval.Nanoseconds()),
		interval:            interval,
	}
	if err := rate.Validate(); err != nil {
		return nil, err
	}
	return rate, nil
}

func (r *Rate) Interval() time.Duration {
	return r.interval
}

func (r *Rate) PerNanosecond() float64 {
	return r.tokensPerNanosecond
}

func (r *Rate) Burst() float64 {
	return r.tokensPerNanosecond * float64(r.interval.Nanoseconds())
}

func (r *Rate) Validate() error {
	if r.tokensPerNanosecond == math.Inf(1) {
		return errors.New("infinite rate")
	}
	if r.tokens == 0 || r.tokensPerNanosecond == 0 {
		return errors.New("zero rate")
	}
	if r.tokens < 0 || r.tokensPerNanosecond < 0 {
		return errors.New("negative rate")
	}
	if r.tokens < 1 {
		return errors.New("limit must be greater than 1")
	}
	if r.tokens > 1<<32 {
		return errors.New("limit is too large")
	}
	if r.interval < time.Millisecond*20 {
		return errors.New("interval must be greater than 20ms")
	}
	if r.interval > time.Hour*24 {
		return errors.New("maximum interval is 24 hours")
	}
	return nil
}

// ReplenishedTokens returns fractional token amount based on time passed.
func (r *Rate) ReplenishedTokens(from, to time.Time) float64 {
	return float64(to.Sub(from).Nanoseconds()) * r.tokensPerNanosecond
}

func (r *Rate) FasterThan(a *Rate) bool {
	return r.tokensPerNanosecond > a.tokensPerNanosecond
}

func (r *Rate) SlowerThan(a *Rate) bool {
	return r.tokensPerNanosecond < a.tokensPerNanosecond
}

func (r *Rate) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Float64("tokens", r.tokens),
		slog.Duration("interval", r.interval),
		slog.Float64("per_second", float64(time.Second.Nanoseconds())*r.tokensPerNanosecond),
	)
}

func (r *Rate) String() string {
	if r == nil {
		return "<nil> rate"
	}
	if r.interval >= time.Hour {
		return fmt.Sprintf(
			"%.2f per %.2f hours",
			r.tokens,
			r.interval.Hours(),
		)
	}
	if r.interval >= time.Minute {
		return fmt.Sprintf(
			"%.2f per %.2f minutes",
			r.tokens,
			r.interval.Minutes(),
		)
	}
	return fmt.Sprintf(
		"%.2f per %.2f seconds",
		r.tokens,
		r.interval.Seconds(),
	)
}
