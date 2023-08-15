package timing

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type unsafeRandomSource struct {
	r *rand.Rand
	b int64
	d int64
}

// UnsafeRandomSourceOption provides configuration for [NewUnsafeRandomSource].
type UnsafeRandomSourceOption func(*unsafeRandomSource) error

func (u *unsafeRandomSource) GetDeviation() (time.Duration, error) {
	return time.Duration(u.b + u.r.Int63n(u.d)), nil
}

func NewUnsafeRandomSource(withOptions ...UnsafeRandomSourceOption) (urs *unsafeRandomSource, err error) {
	urs = &unsafeRandomSource{}
	for _, option := range append(
		withOptions,
		WithDefaultUnsafeRandomSource(),
		WithDefaultUnsafeRandomSourceDeviation(),
	) {
		if err = option(urs); err != nil {
			return nil, fmt.Errorf("cannot initialize unsafe random source: %w", err)
		}
	}
	return urs, nil
}

func WithUnsafeRandom(r *rand.Rand) UnsafeRandomSourceOption {
	return func(u *unsafeRandomSource) error {
		if r == nil {
			return errors.New("cannot use a <nil> random number generator")
		}
		if u.r != nil {
			return errors.New("random number generator is already set")
		}
		u.r = r
		return nil
	}
}

func WithUnsafeRandomSource(s rand.Source) UnsafeRandomSourceOption {
	return func(u *unsafeRandomSource) error {
		if s == nil {
			return errors.New("cannot use a <nil> source of random")
		}
		return WithUnsafeRandom(rand.New(s))(u)
	}
}

func WithDefaultUnsafeRandomSource() UnsafeRandomSourceOption {
	return func(u *unsafeRandomSource) error {
		if u.r != nil {
			return nil // already set
		}
		return WithUnsafeRandomSource(rand.New(rand.NewSource(
			time.Now().UnixNano(),
		)))(u)
	}
}

func WithUnsafeRandomSourceDeviation(d time.Duration) UnsafeRandomSourceOption {
	return func(u *unsafeRandomSource) error {
		if d < MinimumDeviation {
			return fmt.Errorf("timing deviation of %d is smaller than the minimum allowed value of %d", d, MinimumDeviation)
		}
		if u.d != 0 {
			return errors.New("timing deviation is already set")
		}
		u.d = int64(d)
		return nil
	}
}

func WithDefaultUnsafeRandomSourceDeviation() UnsafeRandomSourceOption {
	return func(u *unsafeRandomSource) error {
		if u.d != 0 {
			return nil // already set
		}
		return WithUnsafeRandomSourceDeviation(MinimumDeviation)(u)
	}
}

func WithUnsafeRandomSourceBase(d time.Duration) UnsafeRandomSourceOption {
	return func(u *unsafeRandomSource) error {
		if d == 0 {
			return errors.New("cannot use a zero base deviation")
		}
		if d < 0 {
			return fmt.Errorf("base deviation of %d is less than zero", d)
		}
		if u.b > 0 {
			return errors.New("base deviation is already set")
		}
		u.b = int64(d)
		return nil
	}
}
