/*
Package tagbycontext implements [request.Limiter] by extracting values from request [context.Context].
*/
package tagbycontext

import (
	"fmt"
	"net/http"

	"github.com/dkotik/oakratelimiter/rate"
	"github.com/dkotik/oakratelimiter/request"
)

// ContextLimiter limits HTTP requests by context value associated with a chosen key.
type ContextLimiter struct {
	key     any
	limiter rate.Limiter
	noValue request.Limiter
}

func New(withOptions ...Option) (_ *ContextLimiter, err error) {
	o := &options{}
	for _, option := range append(
		withOptions,
		WithDefaultNoValueLimiter(),
	) {
		if err = option(o); err != nil {
			return nil, fmt.Errorf("cannot initialize context limiter: %w", err)
		}
	}
	return &ContextLimiter{
		key:     o.Key,
		limiter: o.Limiter,
		noValue: o.NoValue,
	}, nil
}

func (c *ContextLimiter) Key() any {
	return c.key
}

func (c *ContextLimiter) Rate() *rate.Rate {
	rate := c.limiter.Rate()
	noValueRate := c.noValue.Rate()
	if rate.SlowerThan(noValueRate) {
		return rate
	}
	return noValueRate
}

func (c *ContextLimiter) Take(
	r *http.Request,
) (
	remaining float64,
	ok bool,
	err error,
) {
	ctx := r.Context()
	value := ctx.Value(c.key)
	if value == nil {
		return c.noValue.Take(r)
	}
	return c.limiter.Take(ctx, fmt.Sprintf("%v", value), 1.0)
}
