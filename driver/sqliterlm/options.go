package sqliterlm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver

	"github.com/dkotik/oakratelimiter/rate"
)

type options struct {
	Database        *sql.DB
	Table           string
	Rate            *rate.Rate
	Burst           float64
	CleanupInterval time.Duration
	CleanupContext  context.Context
}

// Option configures the Postgres rate limiter implementation.
type Option func(*options) error

// WithDatabase provides the Postgres connection for the [RateLimiter].
func WithDatabase(db *sql.DB) Option {
	return func(o *options) error {
		if db == nil {
			return errors.New("cannot use a <nil> database")
		}
		if o.Database != nil {
			return errors.New("database is already set")
		}
		o.Database = db
		return nil
	}
}

// WithDatabaseURL tries to connect to the given database URL.
func WithDatabaseURL(URL string) Option {
	return func(o *options) error {
		if URL == "" {
			return errors.New("cannot use an empty database URL")
		}
		if strings.Index(URL, ":memory:") > 0 {
			slog.Warn("using SQLite in-memory driver, which is slower than mutexrlm")
		}
		db, err := sql.Open("sqlite", URL)
		if err != nil {
			return fmt.Errorf("cannot connect to the %q database: %w", URL, err)
		}
		db.SetMaxOpenConns(1)
		if err = db.Ping(); err != nil {
			return fmt.Errorf("cannot reach the %q database: %w", URL, err)
		}
		go func(ctx context.Context, db *sql.DB) {
			<-ctx.Done()
			_ = db.Close()
		}(context.Background(), db)
		return WithDatabase(db)(o)
	}
}

// WithFile persists rate limiting counters inside a file at the specified path.
func WithFile(path string) Option {
	// fmt.Println(path)
	slog.Debug("rate limiter is using an SQLite3 file", slog.String("path", path))
	return WithDatabaseURL(path)
}

// WithTemporaryFile uses a `oakratelimiter.sqlite3` file inside the system temporary directory. The responsibility for the cleaned up is entrusted to the operating system.
func WithTemporaryFile() Option {
	// if cleanUp == nil {
	// 	cleanUp = context.Background()
	// }
	p := filepath.Join(os.TempDir(), "oakratelimiter.sqlite3")
	// go func(ctx context.Context, p string) {
	// 	<-ctx.Done()
	// 	err := ctx.Err()
	// 	if err != nil && !errors.Is(err, context.Canceled) {
	// 		slog.WarnContext(ctx, "clean up context exited with an error", slog.Any("error", err))
	// 	}
	// 	if err = os.Remove(p); err != nil {
	// 		slog.WarnContext(ctx, "failed to clean up temporary file", slog.String("path", p), slog.Any("error", err))
	// 	}
	// 	fmt.Println("deleted", p)
	// }(cleanUp, p)
	return WithFile(p) // "?cache=shared&mode=rwc"
}

// WithDatabaseFromEnvironment loads [WithDatabaseURL] with value of an environment variable.
func WithDatabaseFromEnvironment(variableName string) Option {
	return func(o *options) (err error) {
		if variableName == "" {
			return errors.New("cannot use an empty environment variable name")
		}
		if err = WithDatabaseURL(os.Getenv(variableName))(o); err != nil {
			return fmt.Errorf("cannot use environment variable %q to create database connection: %w", variableName, err)
		}
		return nil
	}
}

// WithDefaultEphemeralDatabase creates ephemeral RAM database, if no database was provided by another option.
func WithDefaultEphemeralDatabase() Option {
	return func(o *options) (err error) {
		if o.Database != nil {
			return nil // already set
		}
		return WithDatabaseURL(":memory:?cache=shared&mode=rwc")(o)
	}
}

// WithTable specifies the name of the Postgres table that holds token information.
func WithTable(name string) Option {
	return func(o *options) error {
		if !regexp.MustCompile(`^\w+$`).MatchString(name) {
			return fmt.Errorf("table name %q is invalid", name)
		}
		if o.Table != "" {
			return errors.New("table name is already set")
		}
		o.Table = name
		return nil
	}
}

// WithDefaultTable sets [WithTable] to `oakratelimiter`.
func WithDefaultTable() Option {
	return func(o *options) error {
		if o.Table != "" {
			return nil // already set
		}
		return WithTable("oakratelimiter")(o)
	}
}

// WithRate specifies [rate.Rate] setting to use with this rate limiter.
func WithRate(r *rate.Rate) Option {
	return func(o *options) error {
		if r == nil {
			return errors.New("cannot use a <nil> rate")
		}
		if o.Rate != nil {
			return errors.New("rate is already set")
		}
		o.Rate = r
		return nil
	}
}

// WithNewRate applies a new [rate.Rate].
func WithNewRate(limit float64, interval time.Duration) Option {
	return func(o *options) error {
		rate, err := rate.New(limit, interval)
		if err != nil {
			return fmt.Errorf("cannot use new rate: %w", err)
		}
		return WithRate(rate)(o)
	}
}

// WithBurst overrides the default [rate.Rate] burst.
func WithBurst(limit float64) Option {
	return func(o *options) error {
		if limit <= 0 {
			return errors.New("burst limit must be greater than zero")
		}
		if o.Burst != 0 {
			return errors.New("burst limit is already set")
		}
		o.Burst = limit
		return nil
	}
}

// WithDoubleBurst doubles the maximum number of tokens that was last set. Repeat this option to double again.
func WithDoubleBurst() Option {
	return func(o *options) (err error) {
		if err = WithDefaultBurst()(o); err != nil {
			return err
			o.Burst = o.Burst * 2
			return nil
		}
	}
}

// WithTripleBurst triples the maximum number of tokens that was last set. Repeat this option to double again.
func WithTripleBurst() Option {
	return func(o *options) (err error) {
		if err = WithDefaultBurst()(o); err != nil {
			return err
			o.Burst = o.Burst * 3
			return nil
		}
	}
}

// WithDefaultBurst calculates default burst value by counting the maximum number of tokens that can regenerate during the rate interval.
func WithDefaultBurst() Option {
	return func(o *options) error {
		if o.Burst != 0 {
			return nil // already set
		}
		if o.Rate == nil {
			return errors.New("rate is required")
		}
		o.Burst = o.Rate.PerNanosecond() * float64(o.Rate.Interval().Nanoseconds())
		return nil
	}
}

// WithCleanupInterval sets the frequency of map clean up. Lower value frees up more memory at the cost of CPU cycles.
func WithCleanupInterval(of time.Duration) Option {
	return func(o *options) error {
		if o.CleanupInterval != 0 {
			return errors.New("clean up period is already set")
		}
		if of < time.Second {
			return errors.New("clean up period must be greater than 1 second")
		}
		if of > time.Hour {
			return errors.New("clean up period must be less than one hour")
		}
		o.CleanupInterval = of
		return nil
	}
}

// WithDefaultCleanupInterval sets clean up period to 11 minutes.
func WithDefaultCleanupInterval() Option {
	return func(o *options) error {
		if o.CleanupInterval != 0 {
			return nil // already set
		}
		return WithCleanupInterval(time.Minute * 17)(o)
	}
}

// WithCleanupContext controls record clean up cycle. When this context is cancelled, the clean up cycle stops.
func WithCleanupContext(ctx context.Context) Option {
	return func(o *options) error {
		if ctx == nil {
			return fmt.Errorf("cannot use a %q clean up context", ctx)
		}
		if o.CleanupContext != nil {
			return errors.New("clean up context is already set")
		}
		o.CleanupContext = ctx
		return nil
	}
}

// WithDefaultCleanupContext sets clean up context to [context.Background].
func WithDefaultCleanupContext() Option {
	return func(o *options) error {
		if o.CleanupContext != nil {
			return nil // already set
		}
		o.CleanupContext = context.Background()
		return nil
	}
}
