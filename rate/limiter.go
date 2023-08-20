package rate

import (
	"context"
	"errors"
)

// TagFilter directs a [BypassLimiter] to drop tags that return false.
type TagFilter func(tag string) bool

// Limiter contrains the number of consumed tokens to a certain [Rate].
type Limiter interface {
	Rate() *Rate
	Take(
		ctx context.Context,
		tag string,
		tokens float64,
	) (
		remaining float64,
		ok bool,
		err error,
	)
}

type BypassLimiter struct {
	Limiter
	filter TagFilter
}

func NewBypassLimiter(next Limiter, filter TagFilter) (Limiter, error) {
	if next == nil {
		return nil, errors.New("cannot use a <nil> next limiter")
	}
	if filter == nil {
		return nil, errors.New("cannot use a <nil> filter")
	}
	return &BypassLimiter{
		Limiter: next,
		filter:  filter,
	}, nil
}

func NewListBypassLimiter(next Limiter, skipTags ...string) (Limiter, error) {
	if len(skipTags) == 0 {
		return nil, errors.New("cannot use an empty skip list")
	}
	l, err := NewBypassLimiter(next, func(tag string) bool {
		for _, skip := range skipTags {
			if skip == tag {
				return false
			}
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (b *BypassLimiter) Take(
	ctx context.Context,
	tag string,
	tokens float64,
) (
	remaining float64,
	ok bool,
	err error,
) {
	if !b.filter(tag) {
		return tokens, true, nil // drop
	}
	return b.Limiter.Take(ctx, tag, tokens)
}
