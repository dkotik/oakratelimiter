package rate

import "context"

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
