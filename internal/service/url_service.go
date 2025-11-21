package service

import (
	"errors"
	"fmt"

	"github.com/argad/url-shortener/internal/storage"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// URLService provides business logic for URL shortening operations.
// It's used by both HTTP and gRPC handlers as a common layer.
type URLService struct {
	storage storage.Storage
	baseURL string
	logger  *zap.Logger
}

// NewURLService creates a new instance of URLService
func NewURLService(storage storage.Storage, baseURL string, logger *zap.Logger) *URLService {
	return &URLService{
		storage: storage,
		baseURL: baseURL,
		logger:  logger,
	}
}

// ShortenURLResult contains the result of URL shortening operation
type ShortenURLResult struct {
	ShortURL      string
	ShortCode     string
	AlreadyExists bool
}

// ShortenURL shortens a single URL for a given user
func (s *URLService) ShortenURL(userID, originalURL string) (*ShortenURLResult, error) {
	if originalURL == "" {
		return nil, errors.New("URL is required")
	}

	// Generate short ID
	shortID := uuid.New().String()

	// Save to storage - SaveURL returns (string, error)
	urlKey, err := s.storage.SaveURL(originalURL, shortID, userID)
	if err != nil {
		s.logger.Error("Failed to save URL", zap.Error(err))
		return nil, fmt.Errorf("failed to save URL: %w", err)
	}

	return &ShortenURLResult{
		ShortURL:      s.baseURL + "/" + urlKey,
		ShortCode:     urlKey,
		AlreadyExists: false,
	}, nil
}

// BatchItem represents a single item in a batch operation
type BatchItem struct {
	CorrelationID string
	OriginalURL   string
}

// BatchResult represents a result of a batch operation
type BatchResult struct {
	CorrelationID string
	ShortURL      string
}

// ShortenURLBatch shortens multiple URLs in a single operation
func (s *URLService) ShortenURLBatch(userID string, items []BatchItem) ([]BatchResult, error) {
	if len(items) == 0 {
		return []BatchResult{}, nil
	}

	// Convert to storage BatchURLData format
	batchData := make([]storage.BatchURLData, 0, len(items))
	for _, item := range items {
		shortID := uuid.New().String()
		batchData = append(batchData, storage.BatchURLData{
			CorrelationID: item.CorrelationID,
			OriginalURL:   item.OriginalURL,
			ShortURL:      shortID,
		})
	}

	// Save batch to storage - SaveBatch returns ([]BatchURLData, error)
	savedData, err := s.storage.SaveBatch(batchData, userID)
	if err != nil {
		s.logger.Error("Failed to save batch", zap.Error(err))
		return nil, fmt.Errorf("failed to save batch: %w", err)
	}

	// Build results
	results := make([]BatchResult, 0, len(savedData))
	for _, data := range savedData {
		results = append(results, BatchResult{
			CorrelationID: data.CorrelationID,
			ShortURL:      s.baseURL + "/" + data.ShortURL,
		})
	}

	return results, nil
}

// GetOriginalURL retrieves the original URL by short code
func (s *URLService) GetOriginalURL(shortCode string) (string, bool, error) {
	if shortCode == "" {
		return "", false, errors.New("short code is required")
	}

	originalURL, err := s.storage.GetURL(shortCode)
	if err != nil {
		// Check if URL was deleted
		var deletedErr *storage.URLDeletedError
		if errors.As(err, &deletedErr) {
			return "", true, err
		}

		s.logger.Error("Failed to get URL", zap.Error(err))
		return "", false, fmt.Errorf("failed to get URL: %w", err)
	}

	return originalURL, false, nil
}

// UserURL represents a user's URL pair
type UserURL struct {
	ShortURL    string
	OriginalURL string
}

// GetUserURLs retrieves all URLs for a given user
func (s *URLService) GetUserURLs(userID string) ([]UserURL, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	urls, err := s.storage.GetUserURLs(userID)
	if err != nil {
		s.logger.Error("Failed to get user URLs", zap.Error(err))
		return nil, fmt.Errorf("failed to get user URLs: %w", err)
	}

	result := make([]UserURL, len(urls))
	for i, url := range urls {
		result[i] = UserURL{
			ShortURL:    s.baseURL + "/" + url.ShortURL,
			OriginalURL: url.OriginalURL,
		}
	}

	return result, nil
}

// DeleteURLs marks URLs as deleted for a given user
func (s *URLService) DeleteURLs(userID string, shortCodes []string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}
	if len(shortCodes) == 0 {
		return errors.New("short codes are required")
	}

	// Storage.DeleteURLs signature: DeleteURLs(shortURLs []string, userID string) error
	err := s.storage.DeleteURLs(shortCodes, userID)
	if err != nil {
		s.logger.Error("Failed to delete URLs", zap.Error(err))
		return fmt.Errorf("failed to delete URLs: %w", err)
	}

	return nil
}

// GetStats retrieves statistics about URLs and users
func (s *URLService) GetStats() (urlCount int, userCount int, err error) {
	urlCount, err = s.storage.GetURLCount()
	if err != nil {
		s.logger.Error("Failed to get URL count", zap.Error(err))
		return 0, 0, fmt.Errorf("failed to get URL count: %w", err)
	}

	userCount, err = s.storage.GetUserCount()
	if err != nil {
		s.logger.Error("Failed to get user count", zap.Error(err))
		return 0, 0, fmt.Errorf("failed to get user count: %w", err)
	}

	return urlCount, userCount, nil
}
