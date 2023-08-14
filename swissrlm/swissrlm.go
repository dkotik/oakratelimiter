/*
Package swissrlm provides [oakratelimiter.RateLimiter]s that use concurrent [Swiss map] for safe concurrenct access. This strategy is optimal for single-instance rate limiting on large instances.

[Swiss map]: https://github.com/mhmtszr/concurrent-swiss-map
*/
package swissrlm
