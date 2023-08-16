package timing

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"time"
)

type options struct {
	r io.Reader
	d time.Duration
	b time.Duration
}

type Option func(*options) error

func WithRandomByteSource(r io.Reader) Option {
	return func(o *options) error {
		if r == nil {
			return fmt.Errorf("cannot use a %q randon byte source", r)
		}
		if o.r != nil {
			return errors.New("random  byte source is already set")
		}
		o.r = r
		return nil
	}
}

func WithDefaultRandomByteSource() Option {
	return func(o *options) error {
		if o.r != nil {
			return nil // already set
		}
		return WithRandomByteSource(rand.Reader)(o)
	}
}

func WithRandomDeviationLimit(d time.Duration) Option {
	return func(o *options) error {
		if d < MinimumDeviation {
			return fmt.Errorf("timing deviation of %d is smaller than the minimum allowed value of %d", d, MinimumDeviation)
		}
		if o.d != 0 {
			return errors.New("timing deviation is already set")
		}
		o.d = d
		return nil
	}
}

func WithDefaultRandomDeviationLimit() Option {
	return func(o *options) error {
		if o.d != 0 {
			return nil // already set
		}
		return WithRandomDeviationLimit(MinimumDeviation * 2)(o)
	}
}

func WithRandomDeviationBase(d time.Duration) Option {
	return func(o *options) error {
		if d == 0 {
			return errors.New("cannot use a zero base deviation")
		}
		if d < 0 {
			return fmt.Errorf("base deviation of %d is less than zero", d)
		}
		if o.b != 0 {
			return errors.New("base deviation is already set")
		}
		o.b = d
		return nil
	}
}
