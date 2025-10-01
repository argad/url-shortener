package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// URLConflictError is an error that occurs when a URL already exists in the storage.
type URLConflictError struct {
	ExistingShortURL string
}

// Error returns the error message.
func (e *URLConflictError) Error() string {
	return fmt.Sprintf("URL already exists: %s", e.ExistingShortURL)
}

// PostgresStorage is a storage implementation that uses a PostgreSQL database.
type PostgresStorage struct {
	db *pgxpool.Pool
}

// NewPostgresStorage creates a new PostgresStorage.
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
	    user_id VARCHAR(50) NOT NULL,
	    is_deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_original_url ON urls(original_url);
	CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ps.db.Exec(ctx, query)
	return err
}

// SaveURL saves a new URL to the PostgreSQL storage.
func (ps *PostgresStorage) SaveURL(originalURL, shortURL string, userID string) (string, error) {
	if originalURL == "" {
		return "", fmt.Errorf("url cannot be empty")
	}

	query := `
	INSERT INTO urls (short_url, original_url, user_id) 
	VALUES ($1, $2, $3) 
	ON CONFLICT (original_url) 
	DO NOTHING
	RETURNING short_url;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var returnedShortURL string
	err := ps.db.QueryRow(ctx, query, shortURL, originalURL, userID).Scan(&returnedShortURL)
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

// GetURL retrieves a URL from the PostgreSQL storage.
func (ps *PostgresStorage) GetURL(shortURL string) (string, error) {

	query := `
	SELECT original_url, is_deleted FROM urls WHERE short_url = $1;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var url string
	var isDeleted bool
	err := ps.db.QueryRow(ctx, query, shortURL).Scan(&url, &isDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("url with id %s not found", shortURL)
		}

		return "", fmt.Errorf("failed to get URL")
	}

	if isDeleted {
		return "", &URLDeletedError{ShortURL: shortURL}

	}

	return url, nil
}

// SaveBatch saves a batch of URLs to the PostgreSQL storage.
func (ps *PostgresStorage) SaveBatch(batchData []BatchURLData, userID string) ([]BatchURLData, error) {
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

	batch := ps.prepareBatch(batchData, userID)
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

// GetUserURLs retrieves all URLs for a given user ID from the PostgreSQL storage.
func (ps *PostgresStorage) GetUserURLs(userID string) ([]URLData, error) {
	query := `SELECT short_url, original_url, user_id FROM urls WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := ps.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user URLs: %w", err)
	}
	defer rows.Close()

	var userURLs []URLData
	for rows.Next() {
		var urlData URLData
		if err := rows.Scan(&urlData.ShortURL, &urlData.OriginalURL, &urlData.UserID); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		userURLs = append(userURLs, urlData)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return userURLs, nil
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
func (ps *PostgresStorage) prepareBatch(batchData []BatchURLData, userID string) *pgx.Batch {
	batch := &pgx.Batch{}
	query := ps.getBatchInsertQuery()

	for _, item := range batchData {
		batch.Queue(query, item.ShortURL, item.OriginalURL, userID)
	}

	return batch
}

func (ps *PostgresStorage) getBatchInsertQuery() string {
	return `
		INSERT INTO urls (short_url, original_url, user_id) 
		VALUES ($1, $2, $3) 
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

// DeleteURLs deletes a batch of URLs from the PostgreSQL storage.
func (ps *PostgresStorage) DeleteURLs(shortURLs []string, userID string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	batch := &pgx.Batch{}
	query := `UPDATE urls SET is_deleted = true WHERE short_url = $1 AND user_id = $2`

	for _, shortURL := range shortURLs {
		batch.Queue(query, shortURL, userID)
	}

	results := ps.db.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(shortURLs); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("failed to delete URL at index %d: %w", i, err)
		}
	}

	return nil
}
