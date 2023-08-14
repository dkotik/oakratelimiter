/*
Package oakratelimiter protects API endpoints with rate limiting middleware.
*/
package oakratelimiter

import (
	"fmt"
	"net/http"

	"log/slog"
)

type Error interface {
	HyperTextStatusCode() int
}

type Handler interface {
	ServeHyperText(http.ResponseWriter, *http.Request) error
}

type HandlerFunc func(http.ResponseWriter, *http.Request) error

func (f HandlerFunc) ServeHyperText(w http.ResponseWriter, r *http.Request) error {
	return f(w, r)
}

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

type EndpointAccessControl interface {
	IsAllowed(*http.Request) (bool, error)
}

type Middleware struct {
	next         Handler
	headerWriter HeaderWriter
	names        []string
	rateLimiters []RequestLimiter
	errorHandler ErrorHandler
}

func (o *Middleware) ServeHyperText(
	w http.ResponseWriter, r *http.Request,
) (err error) {
	header := w.Header()
	rejected := []string{}
	remaining := float64(0)
	ok := false
	leastRemaining := float64(99999999)
	for i, limiter := range o.rateLimiters {
		remaining, ok, err = limiter.Take(r)
		if leastRemaining > remaining {
			leastRemaining = remaining
		}
		if err != nil {
			o.headerWriter.ReportError(header)
			return fmt.Errorf("rate limiter %q failed: %w", o.names[i], err)
		}
		if !ok {
			rejected = append(rejected, o.names[i])
		}
	}
	if len(rejected) > 0 {
		o.headerWriter.ReportAccessDenied(header, leastRemaining)
		return &TooManyRequestsError{
			rejectedEndpointAccessControlNames: rejected,
		}
	}
	o.headerWriter.ReportAccessAllowed(header, leastRemaining)
	return o.next.ServeHyperText(w, r)
}

func (o *Middleware) ServeHTTP(
	w http.ResponseWriter, r *http.Request,
) {
	o.errorHandler(w, r, o.ServeHyperText(w, r))
}

// TooManyRequestsError indicates overflowing request [Rate].
type TooManyRequestsError struct {
	rejectedEndpointAccessControlNames []string
}

// Error returns a generic text, regardless of what caused the [TooManyRequestsError].
func (e *TooManyRequestsError) Error() string {
	return http.StatusText(http.StatusTooManyRequests)
}

// HTTPStatusCode presents a standard HTTP status code.
func (e *TooManyRequestsError) HTTPStatusCode() int {
	return http.StatusTooManyRequests
}

// LogValue captures causes into structured log entries.
func (e *TooManyRequestsError) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("error", e.Error()),
		slog.Any("rejected_by", e.rejectedEndpointAccessControlNames),
	)
}

// New creates an [Middleware] from either [Basic], [SingleTagging], or [MultiTagging] rate limiters. The selection is based on the [Option]s provided. If the option set contains no request [Tagger]s, [Basic] middleware is returned. If one [Tagger], then [SingleTagging]. If more than one [Tagger], then [MultiTagging]. This function is able to instrument a performant [RateLimiter] for most practical cases.
//
// If you would like more exact or partially obfuscated configuration, use [NewBasic], [NewSingleTagging], [NewMultiTagging] with [NewMiddleware] constructors.
// func New(withOptions ...Option) (Middleware, error) {
// 	o, err := newOptions(append(
// 		withOptions,
// 		func(o *options) error { // validate
// 			return nil
// 		},
// 	)...)
// 	if err != nil {
// 		return nil, fmt.Errorf("cannot create the rate limiter: %w", err)
// 	}
//
// 	if len(o.Tagging) == 0 {
// 		return (&Basic{
// 			failure: NewTooManyRequestsError(
// 				fmt.Errorf("rate limiter %q ran out of tokens", o.Supervising.Name)),
// 			rate:     NewRate(o.Supervising.Limit, o.Supervising.Interval),
// 			limit:    o.Supervising.Limit,
// 			interval: o.Supervising.Interval,
// 			mu:       sync.Mutex{},
// 			bucket:   bucket{},
// 		}).Middleware(), nil
// 	}
//
// 	if o.CleanUpContext == nil {
// 		o.CleanUpContext = context.Background()
// 	}
//
// 	if len(o.Tagging) == 1 {
// 		s := &SingleTagging{
// 			failure: NewTooManyRequestsError(
// 				fmt.Errorf("rate limiter %q ran out of tokens", o.Supervising.Name)),
// 			rate:            NewRate(o.Supervising.Limit, o.Supervising.Interval),
// 			limit:           o.Supervising.Limit,
// 			interval:        o.Supervising.Interval,
// 			mu:              sync.Mutex{},
// 			bucket:          bucket{},
// 			taggedBucketMap: o.Tagging[0],
// 		}
// 		go s.purgeLoop(o.CleanUpContext, o.CleanUpPeriod)
// 		return s.Middleware(), nil
// 	}
//
// 	m := &MultiTagging{
// 		failure: NewTooManyRequestsError(
// 			fmt.Errorf("rate limiter %q ran out of tokens", o.Supervising.Name)),
// 		rate:             NewRate(o.Supervising.Limit, o.Supervising.Interval),
// 		limit:            o.Supervising.Limit,
// 		interval:         o.Supervising.Interval,
// 		mu:               sync.Mutex{},
// 		bucket:           bucket{},
// 		taggedBucketMaps: o.Tagging,
// 	}
// 	go m.purgeLoop(o.CleanUpContext, o.CleanUpPeriod)
// 	return m.Middleware(), nil
// }

// NewMiddleware protects an [Handler] using a [RateLimiter]. The display [Rate] can be used to obfuscate the true [RateLimiter] throughput. HTTP headers are set to promise availability of no more than one call. This is done to conceal the performance capacity of the system, while giving some useful information to API callers regarding service availability. "X-RateLimit-*" headers are experimental, inconsistent in implementation, and meant to be approximate. If display [Rate] is 0, the headers are ommitted.
// func NewMiddleware(l RateLimiter, displayRate Rate) Middleware {
// 	if l == nil {
// 		panic("<nil> rate limiter")
// 	}
//
// 	if displayRate == Rate(0) {
// 		return func(next Handler) Handler {
// 			return func(w http.ResponseWriter, r *http.Request) error {
// 				if err := l.Take(r); err != nil {
// 					return err
// 				}
// 				return next(w, r)
// 			}
// 		}
// 	}
//
// 	limit := uint(1)
// 	oneTokenWindow := time.Nanosecond * time.Duration(1.05/displayRate)
// 	if oneTokenWindow < time.Second {
// 		limit = uint(math.Min(
// 			math.Floor(float64(time.Second.Nanoseconds())*float64(displayRate*0.95)),
// 			1,
// 		))
// 		oneTokenWindow = time.Second
// 	}
// 	displayLimit := fmt.Sprintf("%d", limit)
// 	return func(next Handler) Handler {
// 		return func(w http.ResponseWriter, r *http.Request) error {
// 			t := time.Now().
// 				Add(oneTokenWindow).
// 				UTC().
// 				Format(time.RFC1123)
//
// 			header := w.Header()
// 			header.Set("X-RateLimit-Limit", displayLimit)
// 			header.Set("X-RateLimit-Reset", t)
//
// 			if err := l.Take(r); err != nil {
// 				header.Set("X-RateLimit-Remaining", "0")
// 				header.Set("Retry-After", t)
// 				return err
// 			}
// 			header.Set("X-RateLimit-Remaining", "1")
// 			return next(w, r)
// 		}
// 	}
// }
