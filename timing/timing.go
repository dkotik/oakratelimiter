/*
Package timing provides request modulation middleware that mitigates timing attacks.
*/
package timing

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/dkotik/oakratelimiter"
)

const MinimumDeviation = time.Millisecond

type Source interface {
	GetDeviation() (time.Duration, error)
}

type randomModulator struct {
	next   oakratelimiter.Handler
	source Source
}

func NewMiddleware(source Source) (oakratelimiter.Middleware, error) {
	return func(next oakratelimiter.Handler) oakratelimiter.Handler {
		return &randomModulator{
			next:   next,
			source: source,
		}
	}, nil
}

func (rm *randomModulator) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) error {
	ctx := r.Context()
	wait, err := rm.source.GetDeviation()
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(wait):
		return rm.next.ServeHyperText(w, r)
	}
}

type RandomSource struct {
	r io.Reader
	d *big.Int
	b *big.Int
}

func (r *RandomSource) GetDeviation() (time.Duration, error) {
	d, err := rand.Int(r.r, r.d)
	if err != nil {
		return 0, err
	}
	return time.Duration(r.b.Add(r.b, d).Int64()), nil
}

func NewSource(withOptions ...Option) (_ Source, err error) {
	o := &options{}
	for _, option := range append(
		withOptions,
		WithDefaultRandomByteSource(),
		WithDefaultRandomDeviationLimit(),
	) {
		if err = option(o); err != nil {
			return nil, fmt.Errorf("unable to initiate timing modulator source: %w", err)
		}
	}

	return &RandomSource{
		r: o.r,
		d: big.NewInt(int64(o.d)),
		b: big.NewInt(int64(o.b)),
	}, nil
}

func NewTimingModulator(withOptions ...Option) (oakratelimiter.Middleware, error) {
	source, err := NewSource(withOptions...)
	if err != nil {
		return nil, err
	}
	return NewMiddleware(source)
}
