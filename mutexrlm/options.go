package mutexrlm

import (
	"errors"
	"fmt"
	"time"
)

type options struct {
	Name                  string
	Limit                 float64
	Interval              time.Duration
	InitialAllocationSize int
}

// WithSupervisingLimit sets the top rate limit for either [SingleTagging] or [MultiTagging] rate limiters to prevent [Tagger]s from consuming too much memory. Should be higher than the limit of any request tagger.
func WithSupervisingLimit(withOptions ...Option) Option {
	return func(o *options) (err error) {
		if o.Supervising != nil {
			return errors.New("supervising limit is already set")
		}
		if o.Supervising, err = newLimitOptions(withOptions...); err != nil {
			return fmt.Errorf("cannot create supervising limit: %w", err)
		}
		return nil
	}
}

func newLimitOptions(withOptions ...Option) (*options, error) {
	o := &options{}
	for _, option := range append(
		withOptions,
		func(o *options) error { // validate
			if o.Limit == 0 || o.Interval == 0 {
				return errors.New("WithRate option is required")
			}
			return nil
		},
	) {
		if err := option(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

func newSupervisingLimitOptions(withOptions ...Option) (*options, error) {
	return newLimitOptions(append(
		withOptions,
		WithDefaultName(),
		func(o *options) error {
			if o.InitialAllocationSize != 0 {
				return errors.New("initial allocation size option cannot be applied to the supervising rate limiter")
			}
			return nil
		},
	)...)
}

// Option configures a rate limitter. [Basic] relies on one set of [Option]s. [SingleTagging] and [MultiTagging] use a set for superivising rate limit and additional sets for each [Tagger].
type Option func(*options) error

// WithName associates a name with a rate limiter. It is displayed only in the logs.
func WithName(name string) Option {
	return func(o *options) error {
		if o.Name != "" {
			return errors.New("name has already been set")
		}
		if name == "" {
			return errors.New("cannot use an empty name")
		}
		o.Name = name
		return nil
	}
}

// WithDefaultName sets rate limiter name to "default."
func WithDefaultName() Option {
	return func(o *options) error {
		if o.Name == "" {
			return WithName("default")(o)
		}
		return nil
	}
}

func WithRate(limit float64, interval time.Duration) Option {
	return func(o *options) error {
		if o.Limit != 0 || o.Interval != 0 {
			return errors.New("rate has already been set")
		}
		if limit < 1 {
			return errors.New("limit must be greater than 1")
		}
		if limit > 1<<32 {
			return errors.New("take limit is too large")
		}
		if interval < time.Millisecond*20 {
			return errors.New("interval must be greater than 20ms")
		}
		if interval > time.Hour*24 {
			return errors.New("maximum interval is 24 hours")
		}
		o.Limit = limit
		o.Interval = interval
		return nil
	}
}

// WithInitialAllocationSize sets the number of pre-allocated items for a tagged bucket map. Higher number can improve initial performance at the cost of using more memory.
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

// WithCleanUpPeriod sets the frequency of map clean up. Lower value frees up more memory at the cost of CPU cycles.
func WithCleanUpPeriod(of time.Duration) Option {
	return func(o *options) error {
		if o.CleanUpPeriod != 0 {
			return errors.New("clean up period is already set")
		}
		if of < time.Second {
			return errors.New("clean up period must be greater than 1 second")
		}
		if of > time.Hour {
			return errors.New("clean up period must be less than one hour")
		}
		o.CleanUpPeriod = of
		return nil
	}
}

// WithDefaultCleanUpPeriod sets clean up period to 15 minutes.
func WithDefaultCleanUpPeriod() Option {
	return func(o *options) error {
		if o.CleanUpPeriod == 0 {
			return WithCleanUpPeriod(time.Minute * 15)(o)
		}
		return nil
	}
}
