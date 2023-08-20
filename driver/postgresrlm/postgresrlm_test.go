package postgresrlm

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dkotik/oakratelimiter/test"
)

func TestPostgresDriver(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL is not set")
	}
	rlm, err := New(
		WithDatabaseURL(dbURL),
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
