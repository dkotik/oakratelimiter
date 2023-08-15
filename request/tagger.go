package request

import (
	"context"
	"errors"
	"net"
	"net/http"
)

var Skip = errors.New("skip tagger")

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

// NewIPAddressTagger tags requests by client IP address.
func NewIPAddressTagger() Tagger {
	return func(r *http.Request) (string, error) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return "", err
		}
		return ip, nil
	}
}

// NewIPAddressTagger tags requests by client IP address and port. It is slightly faster than [NewIPAddressTagger].
func NewIPAddressWithPortTagger() Tagger {
	return func(r *http.Request) (string, error) {
		return r.RemoteAddr, nil
	}
}

// NewCookieTagger tags requests by certain cookie value. If [noCookieValue] is an empty string, this tagger issues [SkipTagger].
func NewCookieTagger(name, noCookieValue string) Tagger {
	if noCookieValue == "" {
		return func(r *http.Request) (string, error) {
			cookie, err := r.Cookie(name)
			// if err == http.ErrNoCookie {
			// 	return "", SkipTagger
			// } else
			if err != nil {
				return "", err
			}
			return cookie.Value, nil
		}
	}

	return func(r *http.Request) (string, error) {
		cookie, err := r.Cookie(name)
		if err == http.ErrNoCookie {
			return noCookieValue, nil
		} else if err != nil {
			return "", err
		}
		return cookie.Value, nil
	}
}

// NewHeaderTagger tags requests by certain header value. If [noHeaderValue] is an empty string, this tagger issues [SkipTagger].
// func NewHeaderTagger(name, noHeaderValue string) Tagger {
// 	if noHeaderValue == "" {
// 		return func(r *http.Request) (string, error) {
// 			value := r.Header.Get(name)
// 			if value == "" {
// 				return "", SkipTagger
// 			}
// 			return value, nil
// 		}
// 	}
//
// 	return func(r *http.Request) (string, error) {
// 		value := r.Header.Get(name)
// 		if value == "" {
// 			return noHeaderValue, nil
// 		}
// 		return value, nil
// 	}
// }
