package request

import (
	"context"
	"net/http"
)

// Tagger associates tags to [http.Request]s in order to
// group related requests for a discriminating rate limiter.
// Requests with the same association tag will be tracked
// together by the [Limiter].
type Tagger func(*http.Request) (string, error)

// ContextTagger is used together with [NewRequestTaggerFromContextTagger] to tag requests based on context values. This can help with rate-limiting requests by a role.
type ContextTagger func(context.Context) (string, error)

// NewRequestTaggerFromContextTagger adapts a [ContextTagger] to a [Tagger].
func NewRequestTaggerFromContextTagger(t ContextTagger) Tagger {
	if t == nil {
		panic("cannot use a <nil> context tagger")
	}
	return func(r *http.Request) (string, error) {
		return t(r.Context())
	}
}
