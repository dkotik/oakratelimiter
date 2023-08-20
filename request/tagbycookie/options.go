package tagbycookie

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
	NoCookie request.Limiter
}

type Option func(*options) error

func WithName(name string) Option {
	return func(o *options) error {
		if name == "" {
			return errors.New("cannot use an empty HTTP cookie name")
		}
		if !regexp.MustCompile(`^[a-zA-Z0-9\!\#\$\%\&\'\(\)\*\+\-\.\/\<\>\?\@\[\]\^\_\{\|\}\~]+$`).MatchString(name) {
			return fmt.Errorf("invalid HTTP cookie name: %s", name)
		}
		if o.Name != "" {
			return errors.New("HTTP cookie name is already set")
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

func WithNoCookieLimiter(l request.Limiter) Option {
	return func(o *options) error {
		if l == nil {
			return errors.New("cannot use a <nil> no cookie request limiter")
		}
		if o.NoCookie != nil {
			return errors.New("no cookie request limiter is already set")
		}
		o.NoCookie = l
		return nil
	}
}

func WithNoCookieRateLimiter(tag string, l rate.Limiter) Option {
	return func(o *options) error {
		l, err := request.NewStaticLimiter(tag, l)
		if err != nil {
			return fmt.Errorf("cannot initialize static request limiter: %w", err)
		}
		return WithNoCookieLimiter(l)(o)
	}
}

func WithNoCookieRateLimit(tag string, r *rate.Rate) Option {
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
		return WithNoCookieRateLimiter(tag, l)(o)
	}
}

func WithDefaultNoCookieLimiter() Option {
	return func(o *options) error {
		if o.NoCookie != nil {
			return nil // already set
		}
		if o.Name == "" {
			return errors.New("HTTP cookie name is required")
		}
		if o.Limiter == nil {
			return errors.New("rate limiter is required")
		}
		return WithNoCookieRateLimiter(
			fmt.Sprintf("<cookie %q absent>", o.Name),
			o.Limiter,
		)(o)
	}
}
