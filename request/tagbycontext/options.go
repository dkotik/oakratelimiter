package tagbycontext

import (
	"errors"
	"fmt"
	"time"

	"github.com/dkotik/oakratelimiter/driver/mutexrlm"
	"github.com/dkotik/oakratelimiter/rate"
	"github.com/dkotik/oakratelimiter/request"
)

type options struct {
	Key     any
	Rate    *rate.Rate
	Limiter rate.Limiter
	NoValue request.Limiter
}

type Option func(*options) error

func WithKey(key any) Option {
	return func(o *options) error {
		if key == nil {
			return errors.New("cannot use a <nil> context key")
		}
		if o.Key != nil {
			return errors.New("context key is already set")
		}
		o.Key = key
		return nil
	}
}

func WithRate(r *rate.Rate) Option {
	return func(o *options) (err error) {
		if r == nil {
			return errors.New("cannot use a <nil> rate")
		}
		if o.Rate != nil {
			return errors.New("rate or rate limiter are already set")
		}
		o.Limiter, err = mutexrlm.New(mutexrlm.WithRate(r))
		if err != nil {
			return fmt.Errorf("cannot create rate limiter for rate %q: %w", r.String(), err)
		}
		o.Rate = r
		return nil
	}
}

func WithNewRate(tokens float64, interval time.Duration) Option {
	return func(o *options) error {
		r, err := rate.New(tokens, interval)
		if err != nil {
			return fmt.Errorf("cannot initialize rate: %w", err)
		}
		return WithRate(r)(o)
	}
}

func WithRateLimiter(l rate.Limiter) Option {
	return func(o *options) (err error) {
		if l == nil {
			return errors.New("cannot use a <nil> rate limiter")
		}
		if o.Rate != nil {
			return errors.New("rate or rate limiter are already set")
		}
		r := l.Rate()
		if err = r.Validate(); err != nil {
			return fmt.Errorf("cannot use invalid rate %q: %w", r, err)
		}
		o.Rate = r
		o.Limiter = l
		return nil
	}
}

func WithNoValueLimiter(l request.Limiter) Option {
	return func(o *options) error {
		if l == nil {
			return errors.New("cannot use a <nil> absent context value request limiter")
		}
		if o.NoValue != nil {
			return errors.New("absent context value limiter is already set")
		}
		o.NoValue = l
		return nil
	}
}

func WithNoValueRateLimiter(tag string, l rate.Limiter) Option {
	return func(o *options) error {
		l, err := request.NewStaticLimiter(tag, l)
		if err != nil {
			return fmt.Errorf("cannot initialize static request limiter: %w", err)
		}
		return WithNoValueLimiter(l)(o)
	}
}

func WithNoRateLimit(tag string, r *rate.Rate) Option {
	return func(o *options) (err error) {
		if r == nil {
			return errors.New("cannot use a <nil> rate limit")
		}
		if err = r.Validate(); err != nil {
			return fmt.Errorf("invalid rate limit %q: %w", r, err)
		}
		l, err := mutexrlm.New(mutexrlm.WithRate(r))
		if err != nil {
			return fmt.Errorf("cannot initialize rate limiter: %w", err)
		}
		return WithNoValueRateLimiter(tag, l)(o)
	}
}

func WithDefaultNoValueLimiter() Option {
	return func(o *options) error {
		if o.NoValue != nil {
			return nil // already set
		}
		if o.Key == nil {
			return errors.New("context key is required")
		}
		if o.Limiter == nil {
			return errors.New("rate limiter is required")
		}
		return WithNoValueRateLimiter(
			fmt.Sprintf("<context key %v absent>", o.Key),
			o.Limiter,
		)(o)
	}
}
