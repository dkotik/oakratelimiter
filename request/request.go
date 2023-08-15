/*
Package request defines and implements [Limiter] and [Tagger] for throttling [http.Request]s.
*/
package request

import (
	"errors"
	"net/http"

	"github.com/dkotik/oakratelimiter/rate"
)

type Limiter interface {
	Rate() *rate.Rate
	Take(
		*http.Request,
	) (
		remaining float64,
		ok bool,
		err error,
	)
}

func NewStaticLimiter(tag string, l rate.Limiter) (Limiter, error) {
	if tag == "" {
		return nil, errors.New("cannot use an empty tag")
	}
	if l == nil {
		return nil, errors.New("cannot use a <nil> rate limiter")
	}
	return &staticLimiter{
		tag:     tag,
		limiter: l,
	}, nil
}

type staticLimiter struct {
	tag     string
	limiter rate.Limiter
}

func (s *staticLimiter) Rate() *rate.Rate {
	return s.limiter.Rate()
}

func (s *staticLimiter) Take(r *http.Request) (
	remaining float64,
	ok bool,
	err error,
) {
	return s.limiter.Take(r.Context(), s.tag, 1.0)
}

func NewLimiter(t Tagger, l rate.Limiter) (Limiter, error) {
	if t == nil {
		return nil, errors.New("cannot use a <nil> tagger")
	}
	if l == nil {
		return nil, errors.New("cannot use a <nil> rate limiter")
	}
	return &taggingRequestLimiter{
		tagger:  t,
		limiter: l,
	}, nil
}

type taggingRequestLimiter struct {
	tagger  Tagger
	limiter rate.Limiter
}

func (t *taggingRequestLimiter) Rate() *rate.Rate {
	return t.limiter.Rate()
}

func (t *taggingRequestLimiter) Take(r *http.Request) (
	remaining float64,
	ok bool,
	err error,
) {
	tag, err := t.tagger(r)
	if err != nil {
		return
	}
	return t.limiter.Take(r.Context(), tag, 1.0)
}
