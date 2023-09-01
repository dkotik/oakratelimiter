package oakratelimiter

import (
	"errors"
	"fmt"
	"net/http"

	"log/slog"

	"github.com/dkotik/oakratelimiter/request"
)

// RequestHandler applies a set of [request.Limiter]s to an [http.Request].
type RequestHandler struct {
	next            Handler
	headerWriter    HeaderWriter
	names           []string
	requestLimiters []request.Limiter
}

// ServeHyperText satisfies an improved [http.Handler] interface.
func (rh *RequestHandler) ServeHyperText(
	w http.ResponseWriter, r *http.Request,
) (err error) {
	header := w.Header()
	rejected := []string{}
	remaining := float64(0)
	ok := false
	leastRemaining := float64(99999999)
	for i, limiter := range rh.requestLimiters {
		remaining, ok, err = limiter.Take(r)
		if leastRemaining > remaining {
			leastRemaining = remaining
		}
		if err != nil {
			rh.headerWriter.ReportError(header)
			return fmt.Errorf("rate limiter %q failed: %w", rh.names[i], err)
		}
		if !ok {
			rejected = append(rejected, rh.names[i])
		}
	}
	if len(rejected) > 0 {
		rh.headerWriter.ReportAccessDenied(header, leastRemaining)
		return &TooManyRequestsError{
			rejectedEndpointAccessControlNames: rejected,
		}
	}
	rh.headerWriter.ReportAccessAllowed(header, leastRemaining)
	return rh.next.ServeHyperText(w, r)
}

// ServeHTTP satisfies [http.Handler] for compatibility with the standard library.
func (rh *RequestHandler) ServeHTTP(
	w http.ResponseWriter, r *http.Request,
) {
	err := rh.ServeHyperText(w, r)
	if err == nil {
		return
	}
	var httpError Error
	if errors.As(err, &httpError) {
		msg := err.Error()
		code := httpError.HyperTextStatusCode()
		http.Error(w, msg, code)
		slog.Log(
			r.Context(),
			slog.LevelWarn,
			msg,
			slog.Any("error", err),
			slog.Int("code", code),
		)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
	slog.Log(
		r.Context(),
		slog.LevelError,
		err.Error(),
	)
}

// type SingleLimiterRequestHandler struct {
// 	next           Handler
// 	headerWriter   HeaderWriter
// 	names          []string
// 	requestLimiter request.Limiter
// }
//
// func (rh *SingleLimiterRequestHandler) ServeHyperText(
// 	w http.ResponseWriter, r *http.Request,
// ) (err error) {
// 	header := w.Header()
// 	remaining := float64(0)
// 	ok := false
// 	leastRemaining := float64(99999999)
//
// 	remaining, ok, err = rh.requestLimiter.Take(r)
// 	if leastRemaining > remaining {
// 		leastRemaining = remaining
// 	}
// 	if err != nil {
// 		rh.headerWriter.ReportError(header)
// 		return fmt.Errorf("rate limiter %q failed: %w", rh.names[0], err)
// 	}
// 	if !ok {
// 		rh.headerWriter.ReportAccessDenied(header, leastRemaining)
// 		return &TooManyRequestsError{
// 			rejectedEndpointAccessControlNames: rh.names,
// 		}
// 	}
// 	rh.headerWriter.ReportAccessAllowed(header, leastRemaining)
// 	return rh.next.ServeHyperText(w, r)
// }
//
// func (rh *SingleLimiterRequestHandler) ServeHTTP(
// 	w http.ResponseWriter, r *http.Request,
// ) {
// 	err := rh.ServeHyperText(w, r)
// 	var httpError Error
// 	if errors.As(err, &httpError) {
// 		msg := err.Error()
// 		http.Error(w, msg, httpError.HyperTextStatusCode())
// 		slog.Log(
// 			r.Context(),
// 			slog.LevelError,
// 			msg,
// 			slog.Any("error", httpError),
// 		)
// 		return
// 	}
// 	http.Error(w, err.Error(), http.StatusInternalServerError)
// 	slog.Log(
// 		r.Context(),
// 		slog.LevelError,
// 		err.Error(),
// 	)
// }
