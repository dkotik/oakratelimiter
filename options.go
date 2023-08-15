package oakratelimiter

import (
	"errors"
	"fmt"

	"github.com/dkotik/oakratelimiter/rate"
)

type options struct {
	headerWriter    HeaderWriter
	names           []string
	requestLimiters []rate.RequestLimiter
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
			var least rate.Rate
			for _, l := range o.requestLimiters {
				if least == 0 {
					least = l.Rate()
				}
				if current := l.Rate(); current < least {
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

func (o *options) prepend(name string, l rate.RequestLimiter) error {
	if err := o.isAvailable(name); err != nil {
		return err
	}
	o.names = append([]string{name}, o.names...)
	o.requestLimiters = append([]rate.RequestLimiter{l}, o.requestLimiters...)
	return nil
}

func (o *options) append(name string, l rate.RequestLimiter) error {
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

func WithGlobalRequestLimiter(l rate.RequestLimiter) Option {
	return func(o *options) (err error) {
		if l == nil {
			return fmt.Errorf("cannot use a %q global request limiter", l)
		}
		return o.prepend("global", l)
	}
}

// use mutexrlm.Basic to enforce
// func WithGlobalRateLimiter(rl RateLimiter) Option {
// 	return func(o *options) (err error) {
// 		rl, err := NewBlindRequestLimiter(rl)
// 		if err != nil {
// 			return err
// 		}
// 		return o.append("global", rl)
// 	}
// }
// use mutexrlm.Basic to enforce
// func WithGlobalRateLimit(r Rate) Option {}

func WithRequestLimiter(name string, rl rate.RequestLimiter) Option {
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
		requestLimiter, err := rate.NewRequestLimiter(NewIPAddressTagger(), rl)
		if err != nil {
			return err
		}
		return WithRequestLimiter(
			"internetProtocolAddress",
			requestLimiter,
		)(o)
	}
}

// WithCookieTagger configures rate limiter to track requests based on a certain cookie.
func WithCookieTagger(name string, rl rate.Limiter) Option {
	return func(o *options) (err error) {
		requestLimiter, err := rate.NewRequestLimiter(
			// If [noCookieValue] is an empty string, this [Tagger] issues a [SkipTagger] sentinel value.
			NewCookieTagger(name, ""),
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
