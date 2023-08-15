package oakratelimiter

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/dkotik/oakratelimiter/rate"
)

var _ HeaderWriter = (*SilentHeaderWriter)(nil)
var _ HeaderWriter = (*ObfuscatingHeaderWriter)(nil)

type HeaderWriter interface {
	ReportAccessAllowed(header http.Header, tokens float64)
	ReportAccessDenied(header http.Header, tokens float64)
	ReportError(header http.Header)
}

type SilentHeaderWriter struct{}

func (s *SilentHeaderWriter) ReportAccessAllowed(http.Header, float64) {}
func (s *SilentHeaderWriter) ReportAccessDenied(http.Header, float64)  {}
func (s *SilentHeaderWriter) ReportError(http.Header)                  {}

type ObfuscatingHeaderWriter struct {
	oneTokenWindow   time.Duration
	displayRateLimit string
}

func NewObfuscatingHeaderWriter(r *rate.Rate) HeaderWriter {
	limit := uint(1)
	perNano := r.PerNanosecond()
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
