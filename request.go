package oakratelimiter

import (
	"errors"
	"net/http"
)

type RequestLimiter interface {
	Rate() Rate
	Take(
		*http.Request,
	) (
		remaining float64,
		ok bool,
		err error,
	)
}

// Tagger associates tags to [http.Request]s in order to
// group related requests for a discriminating rate limiter.
// Requests with the same association tag will be tracked
// together by the [RateLimiter]. Return [SkipTagger]
// sentinel value to disregard the [http.Request].
type Tagger func(*http.Request) (string, error)

func NewRequestLimiter(t Tagger, l RateLimiter) (RequestLimiter, error) {
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
	limiter RateLimiter
}

func (t *taggingRequestLimiter) Rate() Rate {
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
