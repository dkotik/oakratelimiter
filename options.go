package oakratelimiter

import (
	"errors"
	"fmt"
)

type options struct {
	next         Handler
	headerWriter HeaderWriter
	names        []string
	global       RequestLimiter
	rateLimiters []RequestLimiter
	errorHandler ErrorHandler
}

func (o *options) push(name string, rl RequestLimiter) error {
	for _, existing := range o.names {
		if existing == name {
			return fmt.Errorf("rate limiter %q is already set", name)
		}
	}
	o.names = append(o.names, name)
	o.rateLimiters = append(o.rateLimiters, rl)
	return nil
}

// Option initializes an [OakRateLimiter] or [Middleware].
type Option func(*options) error

func WithGlobalRequestLimiter(l RequestLimiter) Option {
	return func(o *options) (err error) {
		if l == nil {
			return fmt.Errorf("cannot use a %q global request limiter", l)
		}
		if o.global != nil {
			return errors.New("global request limiter is already set")
		}
		o.global = l
		return nil
	}
}

// use mutexrlm.Basic to enforce
// func WithGlobalRateLimiter(rl RateLimiter) Option {
// 	return func(o *options) (err error) {
// 		rl, err := NewBlindRequestLimiter(rl)
// 		if err != nil {
// 			return err
// 		}
// 		return o.push("global", rl)
// 	}
// }
// use mutexrlm.Basic to enforce
// func WithGlobalRateLimit(r Rate) Option {}

func WithRequestLimiter(name string, rl RequestLimiter) Option {
	return func(o *options) (err error) {
		if rl == nil {
			return errors.New("cannot use a <nil> request limiter")
		}
		return o.push(name, rl)
	}
}

// WithIPAddressTagger configures rate limiter to track requests based on client IP addresses.
func WithIPAddressTagger(rl RateLimiter) Option {
	return func(o *options) (err error) {
		requestLimiter, err := NewRequestLimiter(NewIPAddressTagger(), rl)
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
func WithCookieTagger(name string, rl RateLimiter) Option {
	return func(o *options) (err error) {
		requestLimiter, err := NewRequestLimiter(
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
