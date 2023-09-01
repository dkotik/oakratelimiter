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
	Remaining(
		ctx context.Context,
		tag string,
	) (float64, error)
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

// BypassLimiter uses a [TagFilter] to selectively apply a [Limiter].
type BypassLimiter struct {
	Limiter
	filter TagFilter
}

// NewBypassLimiter attaches a [TagFilter] to a [Limiter].
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

// NewListBypassLimiter creates a [BypassLimiter] that never limits request tags from a given list.
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

// Take consumes tokens, if they are available.
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
		remaining, err = b.Limiter.Remaining(ctx, tag)
		if err != nil {
			return 0, false, err
		}
		return remaining, true, nil // skip tag
	}
	return b.Limiter.Take(ctx, tag, tokens)
}
