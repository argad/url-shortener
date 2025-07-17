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

type URLConflictError struct {
	ExistingShortURL string
}

func (e *URLConflictError) Error() string {
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
			return "", &URLConflictError{ExistingShortURL: existingShortURL}
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
	if err := ps.validateBatchData(batchData); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := ps.beginTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	batch := ps.prepareBatch(batchData)
	results := tx.SendBatch(ctx, batch)

	returnedData, err := ps.processBatchResults(results, batchData)
	if err != nil {
		return nil, err
	}

	results.Close()

	if err := ps.commitTransaction(ctx, tx); err != nil {
		return nil, err
	}

	return returnedData, nil
}

func (ps *PostgresStorage) validateBatchData(batchData []BatchURLData) error {
	if len(batchData) == 0 {
		return fmt.Errorf("batch data cannot be empty")
	}
	return nil
}

func (ps *PostgresStorage) beginTransaction(ctx context.Context) (pgx.Tx, error) {
	tx, err := ps.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

// не слишком мелко?
func (ps *PostgresStorage) prepareBatch(batchData []BatchURLData) *pgx.Batch {
	batch := &pgx.Batch{}
	query := ps.getBatchInsertQuery()

	for _, item := range batchData {
		batch.Queue(query, item.ShortURL, item.OriginalURL)
	}

	return batch
}

func (ps *PostgresStorage) getBatchInsertQuery() string {
	return `
		INSERT INTO urls (short_url, original_url) 
		VALUES ($1, $2) 
		ON CONFLICT (original_url) 
		DO UPDATE SET original_url = EXCLUDED.original_url
		RETURNING short_url;
	`
}

func (ps *PostgresStorage) processBatchResults(results pgx.BatchResults, batchData []BatchURLData) ([]BatchURLData, error) {
	returnedData := make([]BatchURLData, len(batchData))

	for i, item := range batchData {
		returnedShortURL, err := ps.processSingleResult(results, i)
		if err != nil {
			return nil, err
		}

		returnedData[i] = ps.buildBatchResultItem(item, returnedShortURL)
	}

	return returnedData, nil
}

func (ps *PostgresStorage) processSingleResult(results pgx.BatchResults, index int) (string, error) {
	var returnedShortURL string
	err := results.QueryRow().Scan(&returnedShortURL)
	if err != nil {
		return "", fmt.Errorf("failed to save URL at index %d: %w", index, err)
	}
	return returnedShortURL, nil
}

func (ps *PostgresStorage) buildBatchResultItem(originalItem BatchURLData, returnedShortURL string) BatchURLData {
	return BatchURLData{
		CorrelationID: originalItem.CorrelationID,
		OriginalURL:   originalItem.OriginalURL,
		ShortURL:      returnedShortURL,
	}
}

func (ps *PostgresStorage) commitTransaction(ctx context.Context, tx pgx.Tx) error {
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
