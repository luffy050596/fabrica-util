package postgresql

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/go-pantheon/fabrica-util/errors"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Config holds the configuration for PostgreSQL connection
type Config struct {
	DSN             string
	DBName          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
	ConnectTimeout  time.Duration
}

func NewConfig(dsn, dbname string) Config {
	config := DefaultConfig()
	config.DSN = dsn
	config.DBName = dbname

	return config
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		MaxOpenConns:    20,
		MaxIdleConns:    5,
		ConnMaxIdleTime: 15 * time.Minute,
		ConnMaxLifetime: 30 * time.Minute,
		ConnectTimeout:  5 * time.Second,
	}
}

// New creates a new PostgreSQL database connection with the given configuration
func New(driverName string, config Config) (db *sql.DB, cleanup func(), err error) {
	if config.DSN == "" {
		return nil, nil, errors.New("dsn is empty")
	}

	if config.DBName == "" {
		return nil, nil, errors.New("dbname is empty")
	}

	if driverName == "" {
		driverName = "pgx"
	}

	db, err = sql.Open(driverName, config.DSN)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to open database connection")
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}

		return nil, nil, errors.Wrap(err, "failed to ping database")
	}

	cleanup = func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close database connection", "error", err)
		}
	}

	return db, cleanup, nil
}

// NewSimple creates a PostgreSQL connection with simple parameters (backward compatibility)
func NewSimple(dsn, dbname string) (db *sql.DB, cleanup func(), err error) {
	config := DefaultConfig()
	config.DSN = dsn
	config.DBName = dbname

	db, cleanup, err = New("pgx", config)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create database connection")
	}

	return db, cleanup, nil
}

// HealthCheck performs a health check on the database connection
func HealthCheck(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.New("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return errors.Wrap(err, "database health check failed")
	}

	return nil
}

// GetConnectionStats returns connection pool statistics
func GetConnectionStats(db *sql.DB) sql.DBStats {
	if db == nil {
		return sql.DBStats{}
	}

	return db.Stats()
}
