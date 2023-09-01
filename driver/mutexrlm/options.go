package mutexrlm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dkotik/oakratelimiter/rate"
)

type options struct {
	Rate                  *rate.Rate
	Burst                 float64
	InitialAllocationSize int
	CleanupInterval       time.Duration
	CleanupContext        context.Context
}

// Option configures the mutex rate limiter implementation.
type Option func(*options) error

// WithRate sets [rate.Rate] for [RateLimiter].
func WithRate(r *rate.Rate) Option {
	return func(o *options) error {
		if r == nil {
			return errors.New("cannot use a <nil> rate")
		}
		if o.Rate != nil {
			return errors.New("rate is already set")
		}
		o.Rate = r
		return nil
	}
}

// WithNewRate creates a [rate.Rate] to pass to [WithRate] option.
func WithNewRate(limit float64, interval time.Duration) Option {
	return func(o *options) error {
		rate, err := rate.New(limit, interval)
		if err != nil {
			return fmt.Errorf("cannot use new rate: %w", err)
		}
		return WithRate(rate)(o)
	}
}

// WithBurst sets the leaky bucket depth.
func WithBurst(limit float64) Option {
	return func(o *options) error {
		if limit <= 0 {
			return errors.New("burst limit must be greater than zero")
		}
		if o.Burst != 0 {
			return errors.New("burst limit is already set")
		}
		o.Burst = limit
		return nil
	}
}

// WithDefaultBurst applies depth using [rate.Rate] interval to pass to [WithBurst] option. Given a rate of 2 per second, the default burst will be set to 2. Given a rate of 4 per minute, the default burst will be set to 4.
func WithDefaultBurst() Option {
	return func(o *options) error {
		if o.Burst != 0 {
			return nil // already set
		}
		if o.Rate == nil {
			return errors.New("rate is required")
		}
		o.Burst = o.Rate.PerNanosecond() * float64(o.Rate.Interval().Nanoseconds())
		return nil
	}
}

// WithInitialAllocationSize sets the number of pre-allocated items for a tagged bucket map. Higher number can improve starting performance at the cost of using more memory.
func WithInitialAllocationSize(buckets int) Option {
	return func(o *options) error {
		if o.InitialAllocationSize != 0 {
			return errors.New("initial allocation size is already set")
		}
		if buckets < 64 {
			return errors.New("initial allocation size must not be less than 64")
		}
		if buckets > 1<<32 {
			return errors.New("initial allocation size is too great")
		}
		o.InitialAllocationSize = buckets
		return nil
	}
}

// WithDefaultInitialAllocationSize sets initial map allocation to 1024.
func WithDefaultInitialAllocationSize() Option {
	return func(o *options) error {
		if o.InitialAllocationSize == 0 {
			return WithInitialAllocationSize(1024)(o)
		}
		return nil
	}
}

// WithCleanupInterval sets the frequency of map clean up. Lower value frees up more memory at the cost of CPU cycles.
func WithCleanupInterval(of time.Duration) Option {
	return func(o *options) error {
		if o.CleanupInterval != 0 {
			return errors.New("clean up period is already set")
		}
		if of < time.Second {
			return errors.New("clean up period must be greater than 1 second")
		}
		if of > time.Hour {
			return errors.New("clean up period must be less than one hour")
		}
		o.CleanupInterval = of
		return nil
	}
}

// WithDefaultCleanupInterval sets clean up period to 11 minutes.
func WithDefaultCleanupInterval() Option {
	return func(o *options) error {
		if o.CleanupInterval != 0 {
			return nil // already set
		}
		return WithCleanupInterval(time.Minute * 11)(o)
	}
}

// WithCleanupContext provides the [context.Context] for garbage collection. When the context is cancelled, garbage collection stops.
func WithCleanupContext(ctx context.Context) Option {
	return func(o *options) error {
		if ctx == nil {
			return fmt.Errorf("cannot use a %q clean up context", ctx)
		}
		if o.CleanupContext != nil {
			return errors.New("clean up context is already set")
		}
		o.CleanupContext = ctx
		return nil
	}
}

// WithDefaultCleanupContext passes [context.Background] to [WithCleanupContext] option.
func WithDefaultCleanupContext() Option {
	return func(o *options) error {
		if o.CleanupContext != nil {
			return nil // already set
		}
		o.CleanupContext = context.Background()
		return nil
	}
}
