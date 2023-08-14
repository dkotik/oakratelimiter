package oakratelimiter

import (
	"net/http"
	"time"
)

type HeaderWriter interface {
	ReportAccessAllowed(header http.Header, tokens float64)
	ReportAccessDenied(header http.Header, tokens float64)
	ReportError(header http.Header)
}

type ObfuscatingHeaderWriter struct {
	oneTokenWindow   time.Duration
	displayRateLimit string
}

func (o *ObfuscatingHeaderWriter) ReportAccessAllowed(
	h http.Header,
	tokens float64,
) {
	t := time.Now().
		Add(o.oneTokenWindow).
		UTC().
		Format(time.RFC1123)

	h.Set("X-RateLimit-Limit", o.displayRateLimit)
	h.Set("X-RateLimit-Reset", t)
	h.Set("X-RateLimit-Remaining", "1")
}

func (o *ObfuscatingHeaderWriter) ReportAccessDenied(
	h http.Header,
	tokens float64,
) {
	t := time.Now().
		Add(o.oneTokenWindow).
		UTC().
		Format(time.RFC1123)

	h.Set("X-RateLimit-Limit", o.displayRateLimit)
	h.Set("X-RateLimit-Reset", t)
	h.Set("X-RateLimit-Remaining", "0")
	h.Set("Retry-After", t)
}

// ReportError writes appropriate headers for a failing rate limiter.
func (o *ObfuscatingHeaderWriter) ReportError(h http.Header) {
	t := time.Now().
		Add(o.oneTokenWindow).
		UTC().
		Format(time.RFC1123)

	h.Set("X-RateLimit-Limit", o.displayRateLimit)
	h.Set("X-RateLimit-Reset", t)
}
