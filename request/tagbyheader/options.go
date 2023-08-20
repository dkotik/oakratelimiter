package tagbyheader

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/dkotik/oakratelimiter/driver/mutexrlm"
	"github.com/dkotik/oakratelimiter/rate"
	"github.com/dkotik/oakratelimiter/request"
)

type options struct {
	Name     string
	Rate     *rate.Rate
	Limiter  rate.Limiter
	NoHeader request.Limiter
}

type Option func(*options) error

func WithName(name string) Option {
	return func(o *options) error {
		if name == "" {
			return errors.New("cannot use an empty header name")
		}
		if !regexp.MustCompile(`^[a-zA-Z0-9\!\#\$\%\&\'\(\)\*\+\-\.\/\<\>\?\@\[\]\^\_\{\|\}\~]+$`).MatchString(name) {
			return fmt.Errorf("invalid header name: %s", name)
		}
		if o.Name != "" {
			return errors.New("header name is already set")
		}
		o.Name = name
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

func WithNoHeaderLimiter(l request.Limiter) Option {
	return func(o *options) error {
		if l == nil {
			return errors.New("cannot use a <nil> no header request limiter")
		}
		if o.NoHeader != nil {
			return errors.New("no header request limiter is already set")
		}
		o.NoHeader = l
		return nil
	}
}

func WithNoHeaderRateLimiter(tag string, l rate.Limiter) Option {
	return func(o *options) error {
		l, err := request.NewStaticLimiter(tag, l)
		if err != nil {
			return fmt.Errorf("cannot initialize static request limiter: %w", err)
		}
		return WithNoHeaderLimiter(l)(o)
	}
}

func WithNoHeaderRateLimit(tag string, r *rate.Rate) Option {
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
		return WithNoHeaderRateLimiter(tag, l)(o)
	}
}

func WithDefaultNoHeaderLimiter() Option {
	return func(o *options) error {
		if o.NoHeader != nil {
			return nil // already set
		}
		if o.Name == "" {
			return errors.New("header name is required")
		}
		if o.Limiter == nil {
			return errors.New("rate limiter is required")
		}
		return WithNoHeaderRateLimiter(
			fmt.Sprintf("<header %q absent>", o.Name),
			o.Limiter,
		)(o)
	}
}
