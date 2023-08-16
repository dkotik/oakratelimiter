package oakratelimiter

import (
	"errors"
	"fmt"

	"github.com/dkotik/oakratelimiter/driver/mutexrlm"
	"github.com/dkotik/oakratelimiter/rate"
	"github.com/dkotik/oakratelimiter/request"
)

type options struct {
	headerWriter    HeaderWriter
	names           []string
	requestLimiters []request.Limiter
}

func newOptions(from []Option) (o *options, err error) {
	o = &options{}
	for _, option := range append(
		from,
		func(o *options) error { // default header writer
			if o.headerWriter != nil {
				return nil // already set
			}
			if len(o.names) == 0 || len(o.requestLimiters) == 0 {
				return errors.New("at least one request limiter is required")
			}
			var least *rate.Rate
			for _, l := range o.requestLimiters {
				if least == nil {
					least = l.Rate()
				}
				if current := l.Rate(); current.FasterThan(least) {
					least = current
				}
			}
			return WithHeaderWriter(NewObfuscatingHeaderWriter(least))(o)
		},
	) {
		if err = option(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

func (o *options) isAvailable(name string) error {
	for _, existing := range o.names {
		if existing == name {
			return fmt.Errorf("rate limiter %q is already set", name)
		}
	}
	return nil
}

func (o *options) prepend(name string, l request.Limiter) error {
	if err := o.isAvailable(name); err != nil {
		return err
	}
	o.names = append([]string{name}, o.names...)
	o.requestLimiters = append([]request.Limiter{l}, o.requestLimiters...)
	return nil
}

func (o *options) append(name string, l request.Limiter) error {
	if err := o.isAvailable(name); err != nil {
		return err
	}
	o.names = append(o.names, name)
	o.requestLimiters = append(o.requestLimiters, l)
	return nil
}

// Option initializes an [OakRateLimiter] or [Middleware].
type Option func(*options) error

func WithHeaderWriter(h HeaderWriter) Option {
	return func(o *options) error {
		if h == nil {
			return fmt.Errorf("cannot use a %q header writer", h)
		}
		if o.headerWriter != nil {
			return errors.New("header writer is already set")
		}
		o.headerWriter = h
		return nil
	}
}

// WithGlobalRequestLimiter applies [mutexrlm.RequestLimiter] as the top request limiter named "global".
func WithGlobalRequestLimiter(l request.Limiter) Option {
	return func(o *options) (err error) {
		if l == nil {
			return fmt.Errorf("cannot use a %q global request limiter", l)
		}
		return o.prepend("global", l)
	}
}

func WithGlobalRateLimiter(tag string, l rate.Limiter) Option {
	return func(o *options) (err error) {
		rl, err := request.NewStaticLimiter(tag, l)
		if err != nil {
			return err
		}
		return WithGlobalRequestLimiter(rl)(o)
	}
}

func WithGlobalRate(r *rate.Rate) Option {
	return func(o *options) (err error) {
		rl, err := mutexrlm.NewRequestLimiter(mutexrlm.WithRate(r))
		if err != nil {
			return err
		}
		return WithGlobalRequestLimiter(rl)(o)
	}
}

func WithRequestLimiter(name string, rl request.Limiter) Option {
	return func(o *options) (err error) {
		if rl == nil {
			return errors.New("cannot use a <nil> request limiter")
		}
		return o.append(name, rl)
	}
}

// WithIPAddressTagger configures rate limiter to track requests based on client IP addresses.
func WithIPAddressTagger(rl rate.Limiter) Option {
	return func(o *options) (err error) {
		requestLimiter, err := request.NewLimiter(request.NewIPAddressTagger(), rl)
		if err != nil {
			return err
		}
		return WithRequestLimiter(
			"internetProtocolAddress",
			requestLimiter,
		)(o)
	}
}

func WithIPAddressRate(r *rate.Rate) Option {
	return func(o *options) (err error) {
		rl, err := mutexrlm.New(mutexrlm.WithRate(r))
		if err != nil {
			return err
		}
		return WithIPAddressTagger(rl)(o)
	}
}

// WithCookieTagger configures rate limiter to track requests based on a certain cookie.
func WithCookieTagger(name string, rl rate.Limiter) Option {
	return func(o *options) (err error) {
		requestLimiter, err := request.NewLimiter(
			// If [noCookieValue] is an empty string, this [Tagger] issues a [SkipTagger] sentinel value.
			request.NewCookieTagger(name, ""),
			rl,
		)
		if err != nil {
			return err
		}
		return WithRequestLimiter(
			"cookie:"+name,
			requestLimiter,
		)(o)
	}
}

func WithCookieRate(name string, r *rate.Rate) Option {
	return func(o *options) (err error) {
		rl, err := mutexrlm.New(mutexrlm.WithRate(r))
		if err != nil {
			return err
		}
		return WithCookieTagger(name, rl)(o)
	}
}
