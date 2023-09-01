package oakratelimiter

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/dkotik/oakratelimiter/rate"
)

var ( // enforce interface compliance
	_ HeaderWriter = (*SilentHeaderWriter)(nil)
	_ HeaderWriter = (*ObfuscatingHeaderWriter)(nil)
)

// HeaderWriter reports rate limiter state.
type HeaderWriter interface {
	ReportAccessAllowed(header http.Header, tokens float64)
	ReportAccessDenied(header http.Header, tokens float64)
	ReportError(header http.Header)
}

// SilentHeaderWriter does not write any headers.
type SilentHeaderWriter struct{}

func (s *SilentHeaderWriter) ReportAccessAllowed(http.Header, float64) {}
func (s *SilentHeaderWriter) ReportAccessDenied(http.Header, float64)  {}
func (s *SilentHeaderWriter) ReportError(http.Header)                  {}

// ObfuscatingHeaderWriter reports the rate per second regardless of the real [rate.Rate] interval. It can report the rate different from the actual. This is done to avoid leaking the internal state of the system, which may aid the attackers in overcoming the rate limiter.
type ObfuscatingHeaderWriter struct {
	oneTokenWindow   time.Duration
	displayRateLimit string
}

// NewObfuscatingHeaderWriter creates an [ObfuscatingHeaderWriter] using a given rate, which may differ from the actual rate.
func NewObfuscatingHeaderWriter(displayRate *rate.Rate) HeaderWriter {
	limit := uint(1)
	perNano := displayRate.PerNanosecond()
	oneTokenWindow := time.Nanosecond * time.Duration(1.05/perNano)
	if oneTokenWindow < time.Second {
		limit = uint(math.Min(
			math.Floor(float64(time.Second.Nanoseconds())*float64(perNano*0.95)),
			1,
		))
		oneTokenWindow = time.Second
	}
	return &ObfuscatingHeaderWriter{
		oneTokenWindow:   oneTokenWindow,
		displayRateLimit: fmt.Sprintf("%d", limit),
	}
}

// ReportAccessAllowed indicates that the request was not limited.
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

// ReportAccessDenied indicates the request was blocked due to the rate limiter.
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
