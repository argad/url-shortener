package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type ErrURLConflict struct {
	ExistingShortURL string
}

func (e *ErrURLConflict) Error() string {
	return fmt.Sprintf("URL already exists: %s", e.ExistingShortURL)
}

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

	CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_original_url ON urls(original_url);
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
	INSERT INTO urls (short_url, original_url) 
	VALUES ($1, $2) 
	ON CONFLICT (original_url) 
	DO NOTHING
	RETURNING short_url;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var returnedShortURL string
	err := ps.db.QueryRow(ctx, query, shortURL, originalURL).Scan(&returnedShortURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			selectQuery := `SELECT short_url FROM urls WHERE original_url = $1`
			var existingShortURL string
			err := ps.db.QueryRow(ctx, selectQuery, originalURL).Scan(&existingShortURL)
			if err != nil {
				return "", fmt.Errorf("failed to get existing URL: %w", err)
			}
			return "", &ErrURLConflict{ExistingShortURL: existingShortURL}
		}
		return "", fmt.Errorf("failed to save URL: %w", err)
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

func (ps *PostgresStorage) SaveBatch(batchData []BatchURLData) ([]BatchURLData, error) {
	if len(batchData) == 0 {
		return nil, fmt.Errorf("batch data cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := ps.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	query := `
		INSERT INTO urls (short_url, original_url) 
		VALUES ($1, $2) 
		ON CONFLICT (original_url) 
		DO UPDATE SET original_url = EXCLUDED.original_url
		RETURNING short_url;
	`

	for _, item := range batchData {
		batch.Queue(query, item.ShortURL, item.OriginalURL)
	}

	results := tx.SendBatch(ctx, batch)

	returnedData := make([]BatchURLData, len(batchData))
	for i, item := range batchData {
		var returnedShortURL string
		err := results.QueryRow().Scan(&returnedShortURL)
		if err != nil {
			return nil, fmt.Errorf("failed to save URL at index %d: %w", i, err)
		}

		returnedData[i] = BatchURLData{
			CorrelationID: item.CorrelationID,
			OriginalURL:   item.OriginalURL,
			ShortURL:      returnedShortURL,
		}
	}

	results.Close()

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return returnedData, nil
}
