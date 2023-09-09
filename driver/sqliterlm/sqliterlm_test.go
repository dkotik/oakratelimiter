package sqliterlm

import (
	"context"
	"testing"
	"time"

	"github.com/dkotik/oakratelimiter/test"
)

func TestPostgresDriver(t *testing.T) {
	rlm, err := New(
		// WithDatabaseURL(":memory:?cache=shared&mode=rwc"),
		// WithTemporaryFile(),
		WithNewRate(5, time.Second),
		WithCleanupInterval(time.Minute),
	)
	if err != nil {
		t.Fatal("cannot initialize database:", err)
	}

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	// defer cancel()
	test.RateLimiterTest(context.Background(), rlm, 4)(t)
}
