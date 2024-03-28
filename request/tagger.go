package request

import (
	"context"
	"net/http"
)

// TODO: note that request taggers were tricked in other libraries with double values for X-Forwarded header: use Header.Values(string) []string to overcome, see example below.

/*
Account Takeover Via Rate-Limit Bypass

When a password reset was initiated, users were required to enter a six-digit numeric code sent to their email for verification.

To prevent brute force attacks, the application implemented rate-limit protection, restricting the number of requests users could make within a specific timeframe. If this limit was surpassed, the system issued a 429 Too Many Requests error message.

However, the rate-limit protection was bypassed by adding two X-Forwarded-For: IP headers:

    X-Forwarded-For: 1.1.1.1
    X-Forwarded-For: 2.2.2.2

By replacing the IP address in the second X-Forwarded-For header, it became possible to bypass the rate-limit and try multiple codes until the correct one was found.

Exploiting this vulnerability allowed for the unauthorized takeover of any account within the application.
*/

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
