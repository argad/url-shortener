package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

type Database struct {
	pool *pgxpool.Pool
}

func NewDatabase(databaseDSN string) (*Database, error) {
	if databaseDSN == "" {
		return nil, fmt.Errorf("database DSN is required")
	}

	dbpool, err := pgxpool.New(context.Background(), databaseDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	return &Database{pool: dbpool}, nil
}

func (db *Database) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

func (db *Database) Ping(ctx context.Context) error {
	if db.pool == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	return db.pool.Ping(ctx)
}

func (db *Database) GetPool() *pgxpool.Pool {
	return db.pool
}
