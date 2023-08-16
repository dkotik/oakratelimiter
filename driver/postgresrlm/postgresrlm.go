/*
Package postgresrlm implements [rate.Limiter].
*/
package postgresrlm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"log/slog"

	"github.com/dkotik/oakratelimiter/rate"
)

// RateLimiter keep leaky bucket state in a Postgres database.
type RateLimiter struct {
	rate         *rate.Rate
	burstLimit   float64
	db           *sql.DB
	createStmt   *sql.Stmt
	retrieveStmt *sql.Stmt
	updateStmt   *sql.Stmt
	cleanupStmt  *sql.Stmt
}

func New(withOptions ...Option) (r *RateLimiter, err error) {
	o := &options{}
	for _, option := range append(
		withOptions,
		WithDefaultTable(),
		WithDefaultDatabaseFromEnvironment(),
		WithDefaultBurst(),
		WithDefaultCleanupInterval(),
		WithDefaultCleanupContext(),
		func(o *options) (err error) {
			_, err = o.Database.Exec(`
        CREATE TABLE IF NOT EXISTS
        ` + o.Table + ` (
          tag varchar(128) NOT NULL,
          touched bigint NOT NULL,
          tokens numeric NOT NULL,
          PRIMARY KEY(tag)
        )
      `)
			if err != nil {
				return fmt.Errorf("cannot create database table %q: %w", o.Table, err)
			}
			return nil
		},
	) {
		if err = option(o); err != nil {
			return nil, fmt.Errorf("cannot initialize Postgres rate limiter driver: %w", err)
		}
	}

	r = &RateLimiter{
		rate:       o.Rate,
		burstLimit: o.Burst,
		db:         o.Database,
	}
	r.createStmt, err = r.db.Prepare(`INSERT INTO ` + o.Table + `(touched, tokens, tag) VALUES($1, $2, $3)`)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare create statement: %w", err)
	}
	r.retrieveStmt, err = r.db.Prepare(`SELECT touched, tokens FROM ` + o.Table + ` WHERE tag=$1`)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare retrieve statement: %w", err)
	}
	r.updateStmt, err = r.db.Prepare(`UPDATE ` + o.Table + ` SET touched=$1, tokens=$2 WHERE tag=$3`)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare update statement: %w", err)
	}
	r.cleanupStmt, err = r.db.Prepare(`DELETE FROM ` + o.Table + ` WHERE touched < $1`)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare delete statement: %w", err)
	}

	go func(ctx context.Context, r *RateLimiter, every time.Duration) {
		t := time.NewTicker(every)
		defer t.Stop()
		at := time.Now()
		for {
			if err := r.Cleanup(ctx, at); err != nil {
				slog.Warn(
					"could not clean up expired rate limiter records",
					slog.Any("error", err),
				)
			}
			select {
			case <-ctx.Done():
				return
			case at = <-t.C:
				// continue
			}
		}
	}(o.CleanupContext, r, o.CleanupInterval)

	return r, nil
}

func (r *RateLimiter) Rate() *rate.Rate {
	return r.rate
}

// Take retrieves available tokens by tag and takes one token from it.
func (r *RateLimiter) Take(
	ctx context.Context,
	tag string,
	tokens float64,
) (
	remaining float64,
	ok bool,
	err error,
) {
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	t := time.Now()
	var touched int64
	retrieve := tx.StmtContext(ctx, r.retrieveStmt)
	row := retrieve.QueryRow(tag)
	if err = row.Err(); err != nil {
		return 0, false, err
	}
	if err = row.Scan(&touched, &remaining); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if r.burstLimit < tokens {
				return r.burstLimit, false, nil
			}
			remaining = r.burstLimit - tokens
			_, err = tx.StmtContext(ctx, r.createStmt).Exec(
				t.UnixMicro(),
				remaining,
				tag,
			)
			if err != nil {
				return 0, false, err
			}
			if err = tx.Commit(); err != nil {
				return 0, false, err
			}
			return remaining, true, nil
		}
		return 0, false, err
	}

	remaining += r.rate.ReplenishedTokens(time.UnixMicro(touched), t)
	if remaining > r.burstLimit {
		remaining = r.burstLimit
	}
	if remaining < tokens {
		return remaining, false, nil
	}
	remaining -= tokens
	update := tx.StmtContext(ctx, r.updateStmt)
	if _, err = update.Exec(t.UnixMicro(), remaining, tag); err != nil {
		return 0, false, err
	}

	if err = tx.Commit(); err != nil {
		return 0, false, err
	}
	return remaining, true, nil
}

// Cleanup removes all tokens that are expired by given [time.Time].
func (r *RateLimiter) Cleanup(ctx context.Context, at time.Time) error {
	_, err := r.cleanupStmt.ExecContext(ctx, at.Add(-r.rate.Interval()).UnixMicro())
	// fmt.Println("ran cleap up")
	return err
}
