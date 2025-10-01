package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Database represents a connection to the database.
type Database struct {
	pool *pgxpool.Pool
}

// NewDatabase creates a new database connection.
func NewDatabase(databaseDSN string) (*Database, error) {
	if databaseDSN == "" {
		return nil, fmt.Errorf("database DSN is required")
	}

	dbpool, err := pgxpool.New(context.Background(), databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	return &Database{pool: dbpool}, nil
}

// Close closes the database connection.
func (db *Database) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// Ping pings the database to check if the connection is alive.
func (db *Database) Ping(ctx context.Context) error {
	if db.pool == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	return db.pool.Ping(ctx)
}

// GetPool returns the underlying database connection pool.
func (db *Database) GetPool() *pgxpool.Pool {
	return db.pool
}
