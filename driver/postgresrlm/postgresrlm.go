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
	rate            *rate.Rate
	microSecondRate float64
	burstLimit      float64
	db              *sql.DB
	createStmt      *sql.Stmt
	retrieveStmt    *sql.Stmt
	// updateStmt   *sql.Stmt
	// upsertStmt  *sql.Stmt
	cleanupStmt *sql.Stmt
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
			// _, err = o.Database.Exec(`DROP TABLE IF EXISTS ` + o.Table)
			// if err != nil {
			// 	return err
			// }
			_, err = o.Database.Exec(`
        CREATE TABLE IF NOT EXISTS
        ` + o.Table + ` (
          tag varchar(128) NOT NULL,
          touched bigint NOT NULL,
          tokens numeric NOT NULL
        )
      `)
			if err != nil {
				return fmt.Errorf("cannot create database table %q: %w", o.Table, err)
			}
			_, err = o.Database.Exec(`
        CREATE INDEX IF NOT EXISTS
         oakratelimiterTagIndexFor_` + o.Table + ` ON ` + o.Table + `(tag)`)
			if err != nil {
				return fmt.Errorf("cannot create database index for table %q: %w", o.Table, err)
			}
			return nil
		},
	) {
		if err = option(o); err != nil {
			return nil, fmt.Errorf("cannot initialize Postgres rate limiter driver: %w", err)
		}
	}

	r = &RateLimiter{
		rate:            o.Rate,
		microSecondRate: o.Rate.PerNanosecond() * 1000,
		burstLimit:      o.Burst,
		db:              o.Database,
	}
	// r.upsertStmt, err = r.db.Prepare(fmt.Sprintf(
	// 	// TODO: this upsert is easily DDOS-able, because -$4 tokens are substracted regardless of whether they are available or not.
	// 	// could fix it with a on-row-update trigger?
	// 	`
	//   WITH
	//   original AS (
	//     SELECT tokens FROM %[1]s WHERE tag=$1
	//     UNION SELECT %.6[2]f
	//   ), updated AS (
	//     INSERT INTO %[1]s (tag, touched, tokens)
	//        VALUES($1, $2, %.6[2]f-$4)
	//      ON CONFLICT (tag) DO
	//        UPDATE
	//         SET touched=$2, tokens=GREATEST(LEAST(
	//     		  %[1]s.tokens + (($2-%[1]s.touched)::numeric*$3),
	//     		  %.6[2]f
	//     		)-$4, 0)
	//      RETURNING %[1]s.tokens
	//   )
	//
	//   SELECT
	//     updated.tokens, updated.tokens=original.tokens
	//   FROM original, updated;
	//   `,
	// 	o.Table,
	// 	o.Burst,
	// ))
	// if err != nil {
	// 	return nil, fmt.Errorf("cannot prepare upsert statement: %w", err)
	// }
	r.createStmt, err = r.db.Prepare(`INSERT INTO ` + o.Table + `(tag, touched, tokens) VALUES($1, $2, $3)`)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare create statement: %w", err)
	}
	r.retrieveStmt, err = r.db.Prepare(`SELECT SUM(tokens) FROM ` + o.Table + ` WHERE tag=$1 AND touched>$2`)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare retrieve statement: %w", err)
	}
	// r.updateStmt, err = r.db.Prepare(`UPDATE ` + o.Table + ` SET touched=$1, tokens=$2 WHERE tag=$3`)
	// if err != nil {
	// 	return nil, fmt.Errorf("cannot prepare update statement: %w", err)
	// }
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, false, err
	}
	t := time.Now()

	create := tx.Stmt(r.createStmt)
	_, err = create.Exec(tag, t.UnixMicro(), tokens)
	if err != nil {
		return 0, false, err
	}

	retrieve := tx.Stmt(r.retrieveStmt)
	row := retrieve.QueryRow(tag, t.Add(-r.rate.Interval()).UnixMicro())
	if err = row.Err(); err != nil {
		return 0, false, errors.Join(err, tx.Rollback())
	}
	if err = row.Scan(&remaining); err != nil {
		return 0, false, errors.Join(err, tx.Rollback())
	}
	remaining = r.burstLimit - remaining
	if remaining < 0 { // not enough
		// panic("not enough")
		return remaining, false, tx.Rollback()
	}
	if err = tx.Commit(); err != nil {
		return remaining, false, errors.Join(err, tx.Rollback())
	}
	return remaining, true, nil
}

// Cleanup removes all tokens that are expired by given [time.Time].
func (r *RateLimiter) Cleanup(ctx context.Context, at time.Time) error {
	_, err := r.cleanupStmt.ExecContext(ctx, at.Add(-r.rate.Interval()).UnixMicro())
	return err
}
