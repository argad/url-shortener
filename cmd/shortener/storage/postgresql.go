package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type PostgresStorage struct {
	db *pgxpool.Pool
}

func NewPostgresStorage(db *pgxpool.Pool) (*PostgresStorage, error) {
	storage := &PostgresStorage{db: db}

	if err := storage.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return storage, nil
}

func (ps *PostgresStorage) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		short_url VARCHAR(255) UNIQUE NOT NULL,
		original_url VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ps.db.Exec(ctx, query)
	return err
}

func (ps *PostgresStorage) SaveURL(originalURL, shortURL string) (string, error) {
	if originalURL == "" {
		return "", fmt.Errorf("url cannot be empty")
	}

	query := `
	INSERT INTO urls (short_url, original_url) VALUES ($1, $2) RETURNING short_url;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var returnedShortURL string
	err := ps.db.QueryRow(ctx, query, shortURL, originalURL).Scan(&returnedShortURL)
	if err != nil {

		return "", fmt.Errorf("failed to save URL")
	}

	return returnedShortURL, nil
}

func (ps *PostgresStorage) GetURL(shortURL string) (string, error) {

	query := `
	SELECT original_url FROM urls WHERE short_url = $1;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var url string
	err := ps.db.QueryRow(ctx, query, shortURL).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("url with id %s not found", shortURL)
		}

		return "", fmt.Errorf("failed to get URL")
	}

	return url, nil
}
