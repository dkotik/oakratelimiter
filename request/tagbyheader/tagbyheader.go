/*
Package tagbyheader implements [request.Limiter] by extracting an HTTP header value from [http.Request]s.
*/
package tagbyheader

import (
	"fmt"
	"net/http"

	"github.com/dkotik/oakratelimiter/rate"
	"github.com/dkotik/oakratelimiter/request"
)

type HeaderLimiter struct {
	name     string
	limiter  rate.Limiter
	noHeader request.Limiter
}

func New(withOptions ...Option) (_ *HeaderLimiter, err error) {
	o := &options{}
	for _, option := range append(
		withOptions,
		WithDefaultNoHeaderLimiter(),
	) {
		if err = option(o); err != nil {
			return nil, fmt.Errorf("cannot initialize cookie limiter: %w", err)
		}
	}
	return &HeaderLimiter{
		name:     o.Name,
		limiter:  o.Limiter,
		noHeader: o.NoHeader,
	}, nil
}

func (h *HeaderLimiter) Name() string {
	return h.name
}

func (h *HeaderLimiter) Rate() *rate.Rate {
	rate := h.limiter.Rate()
	noHeaderRate := h.noHeader.Rate()
	if rate.SlowerThan(noHeaderRate) {
		return rate
	}
	return noHeaderRate
}

func (h *HeaderLimiter) Take(
	r *http.Request,
) (
	remaining float64,
	ok bool,
	err error,
) {
	value := r.Header.Get(h.name)
	if value == "" {
		return h.noHeader.Take(r)
	}
	return h.limiter.Take(r.Context(), value, 1.0)
}
