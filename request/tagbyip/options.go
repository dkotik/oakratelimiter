package tagbyip

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/dkotik/oakratelimiter/driver/mutexrlm"
	"github.com/dkotik/oakratelimiter/rate"
)

type AddressExtractor func(*http.Request) (string, error)

type options struct {
	Extractor AddressExtractor
	Limiter   rate.Limiter
	Filter    rate.TagFilter
	Skip      []string
}

type Option func(*options) error

func WithAddressExtractor(e AddressExtractor) Option {
	return func(o *options) error {
		if e == nil {
			return errors.New("cannot use a <nil> address extractor")
		}
		if o.Extractor != nil {
			return errors.New("address extractor is already set")
		}
		o.Extractor = e
		return nil
	}
}

func WithAddressAndPortExtractor() Option {
	return WithAddressExtractor(func(r *http.Request) (string, error) {
		return r.RemoteAddr, nil
	})
}

func WithDefaultAddressExtractor() Option {
	return func(o *options) (err error) {
		if o.Extractor != nil {
			return nil // already set
		}
		return WithAddressExtractor(func(r *http.Request) (string, error) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				return "", err
			}
			return ip, nil
		})(o)
	}
}

func WithRate(r *rate.Rate) Option {
	return func(o *options) (err error) {
		if r == nil {
			return errors.New("cannot use a <nil> rate")
		}
		o.Limiter, err = mutexrlm.New(mutexrlm.WithRate(r))
		if err != nil {
			return fmt.Errorf("cannot create rate limiter for rate %q: %w", r.String(), err)
		}
		return nil
	}
}

func WithNewRate(tokens float64, interval time.Duration) Option {
	return func(o *options) error {
		r, err := rate.New(tokens, interval)
		if err != nil {
			return fmt.Errorf("cannot initialize rate: %w", err)
		}
		return WithRate(r)(o)
	}
}

func WithRateLimiter(l rate.Limiter) Option {
	return func(o *options) (err error) {
		if l == nil {
			return errors.New("cannot use a <nil> rate limiter")
		}
		r := l.Rate()
		if err = r.Validate(); err != nil {
			return fmt.Errorf("cannot use invalid rate %q: %w", r, err)
		}
		o.Limiter = l
		return nil
	}
}

func WithFilter(f rate.TagFilter) Option {
	return func(o *options) error {
		if f == nil {
			return errors.New("cannot use a <nil> tag filter")
		}
		if o.Filter != nil {
			return errors.New("tag filter is already set")
		}
		o.Filter = f
		return nil
	}
}

func WithSkipList(addresses ...string) Option {
	return func(o *options) (err error) {
		if len(addresses) == 0 {
			return errors.New("cannot use an empty skip list")
		}
		for _, address := range addresses {
			for _, current := range o.Skip {
				if address == current {
					return fmt.Errorf("tag %q is already on the skip list", current)
				}
			}
		}
		o.Skip = append(o.Skip, addresses...)
		return nil
	}
}
