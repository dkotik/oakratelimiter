/*
Package tagbycookie implements [request.Limiter] by extracting an HTTP cookie from [http.Request]s.
*/
package tagbycookie

import (
	"fmt"
	"net/http"

	"github.com/dkotik/oakratelimiter/rate"
	"github.com/dkotik/oakratelimiter/request"
)

type CookieLimiter struct {
	name     string
	limiter  rate.Limiter
	noCookie request.Limiter
}

func New(withOptions ...Option) (_ *CookieLimiter, err error) {
	o := &options{}
	for _, option := range append(
		withOptions,
		WithDefaultNoCookieLimiter(),
	) {
		if err = option(o); err != nil {
			return nil, fmt.Errorf("cannot initialize cookie limiter: %w", err)
		}
	}
	return &CookieLimiter{
		name:     o.Name,
		limiter:  o.Limiter,
		noCookie: o.NoCookie,
	}, nil
}

func (c *CookieLimiter) Name() string {
	return c.name
}

func (c *CookieLimiter) Rate() *rate.Rate {
	rate := c.limiter.Rate()
	noCookieRate := c.noCookie.Rate()
	if rate.SlowerThan(noCookieRate) {
		return rate
	}
	return noCookieRate
}

func (c *CookieLimiter) Take(
	r *http.Request,
) (
	remaining float64,
	ok bool,
	err error,
) {
	cookie, err := r.Cookie(c.name)
	switch {
	case cookie == nil || cookie.Value == "":
		return c.noCookie.Take(r)
	case err != nil:
		return
	}
	return c.limiter.Take(r.Context(), cookie.Value, 1.0)
}
