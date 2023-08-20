/*
Package tagbyip implements [request.Limiter] by extracting Internet Protocol addresses from [http.Request]s.
*/
package tagbyip

import (
	"fmt"
	"net/http"

	"github.com/dkotik/oakratelimiter/rate"
	"github.com/dkotik/oakratelimiter/request"
)

type IPAddressLimiter struct {
	extractor AddressExtractor
	limiter   rate.Limiter
	filter    rate.TagFilter
}

func New(withOptions ...Option) (_ request.Limiter, err error) {
	o := &options{}
	for _, option := range append(
		withOptions,
		WithDefaultAddressExtractor(),
		func(o *options) error {
			if len(o.Skip) == 0 {
				return nil
			}
			skipMap := make(map[string]struct{})
			for _, skip := range o.Skip {
				skipMap[skip] = struct{}{}
			}
			if o.Filter == nil {
				o.Filter = func(tag string) bool {
					_, ok := skipMap[tag]
					return !ok
				}
			}
			additionalFilter := o.Filter
			o.Filter = func(tag string) bool {
				if _, ok := skipMap[tag]; ok {
					return false
				}
				return additionalFilter(tag)
			}
			return nil
		},
	) {
		if err = option(o); err != nil {
			return nil, fmt.Errorf("cannot initialize cookie limiter: %w", err)
		}
	}
	if o.Filter == nil {
		o.Filter = func(tag string) bool {
			return true // accept all tags
		}
	}

	return &IPAddressLimiter{
		extractor: o.Extractor,
		limiter:   o.Limiter,
		filter:    o.Filter,
	}, nil
}

func (a *IPAddressLimiter) Rate() *rate.Rate {
	return a.limiter.Rate()
}

func (a *IPAddressLimiter) Take(
	r *http.Request,
) (
	remaining float64,
	ok bool,
	err error,
) {
	address, err := a.extractor(r)
	if err != nil {
		return 0, false, err
	}
	if !a.filter(address) {
		return a.limiter.Rate().Burst(), true, nil
	}
	return a.limiter.Take(r.Context(), address, 1.0)
}
