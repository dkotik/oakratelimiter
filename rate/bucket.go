package rate

import "time"

// LeakyBucket keeps track of available tokens according to a given [Rate].
type LeakyBucket struct {
	touched time.Time
	tokens  float64
}

// NewLeakyBucket returns a full [LeakyBucket].
func NewLeakyBucket(at time.Time, r *Rate, burstLimit float64) *LeakyBucket {
	b := &LeakyBucket{}
	b.Refill(at, r, burstLimit)
	return b
}

// Touched returns the last time the bucket was updated.
func (l *LeakyBucket) Touched() time.Time {
	return l.touched
}

// Refill calculates and returns tokens since last update to given time at a [Rate]. Restored tokens will not exceed the burst limit.
func (l *LeakyBucket) Refill(at time.Time, r *Rate, burstLimit float64) {
	if l.tokens < burstLimit {
		l.tokens = l.tokens + r.ReplenishedTokens(l.touched, at)
		if l.tokens > burstLimit {
			l.tokens = burstLimit
		}
		l.touched = at
	}
}

// Remaining returns the number of tokens in the bucket. Use only after running [LeakyBucket.Refill].
func (l *LeakyBucket) Remaining() float64 {
	return l.tokens
}

// Take removes tokens from the bucket, if that many are available. Use only after running [LeakyBucket.Refill].
func (l *LeakyBucket) Take(tokens float64) (remaining float64, ok bool) {
	if l.tokens < tokens {
		return l.tokens, false
	}
	l.tokens -= tokens
	return l.tokens, true
}

// // bucket tracks remaining tokens and limit expiration.
// type bucket struct {
// 	expires time.Time
// 	tokens  float64
// }
//
// // Expires returns true if the bucket is expired at given [time.Time].
// func (b *bucket) Expires(at time.Time) bool {
// 	return b.expires.Before(at)
// }
//
// // Take removes one token from the bucket. If the bucket is fresh, some fractional amount of tokens is also replenished according to [rate.Rate] over the transpired time since the previous take from the bucket. Must run inside mutex lock.
// func (b *bucket) Take(limit float64, r rate.Rate, from, to time.Time) bool {
// 	if b.Expires(from) { // reset
// 		b.tokens = limit - 1
// 		b.expires = to
// 		return true
// 	}
//
// 	replenished := b.tokens + r.ReplenishedTokens(b.expires, to)
// 	b.expires = to
// 	if replenished < 1 { // nothing to take
// 		b.tokens = replenished
// 		return false
// 	}
//
// 	b.tokens = replenished - 1
// 	return true
// }
