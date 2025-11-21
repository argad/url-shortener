package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/argad/url-shortener/internal/config"
	"github.com/argad/url-shortener/internal/database"
	"github.com/argad/url-shortener/internal/storage"
	"go.uber.org/zap"
)

// ShortenerService encapsulates the business logic of the URL shortener.
type ShortenerService struct {
	storage storage.Storage
	config  *config.Config
	db      *database.Database
	logger  *zap.Logger
}

// ShortenRequest defines the structure for a request to shorten a URL.
// It contains the URL that needs to be shortened.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse defines the structure for the response of a URL shortening request.
// It contains the resulting shortened URL.
type ShortenResponse struct {
	Result string `json:"result"`
}

// BatchURLRequest defines the structure for a single URL in a batch shortening request.
// It includes a correlation ID to track the request and the original URL to be shortened.
type BatchURLRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchURLResponse defines the structure for a single URL in the response of a batch shortening request.
// It includes the correlation ID from the request and the resulting shortened URL.
type BatchURLResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// NewShortenerService creates a new ShortenerService.
func NewShortenerService(
	storage storage.Storage,
	cfg *config.Config,
	db *database.Database,
	logger *zap.Logger,
) *ShortenerService {
	return &ShortenerService{
		storage: storage,
		config:  cfg,
		db:      db,
		logger:  logger,
	}
}

// GetURL retrieves the original URL for a given short URL.
func (s *ShortenerService) GetURL(shortURL string) (string, error) {
	if shortURL == "" {
		return "", errors.New("short URL cannot be empty")
	}

	url, err := s.storage.GetURL(shortURL)
	if err != nil {
		return "", err
	}

	return url, nil
}

// ShortenURL creates a new shortened URL.
func (s *ShortenerService) ShortenURL(originalURL, userID string) (string, error) {
	if !strings.HasPrefix(originalURL, "http://") && !strings.HasPrefix(originalURL, "https://") {
		return "", errors.New("invalid URL format")
	}

	id := generateID()
	urlKey, err := s.storage.SaveURL(originalURL, id, userID)
	if err != nil {
		return "", err
	}

	return urlKey, nil
}

// ShortenBatch processes a batch of URLs for shortening.
func (s *ShortenerService) ShortenBatch(batchReq []BatchURLRequest, userID string) ([]BatchURLResponse, error) {
	batchData := make([]storage.BatchURLData, len(batchReq))

	for i, item := range batchReq {
		if item.OriginalURL == "" {
			return nil, errors.New("URL cannot be empty in batch item")
		}
		if item.CorrelationID == "" {
			return nil, errors.New("correlation ID cannot be empty in batch item")
		}
		if !strings.HasPrefix(item.OriginalURL, "http://") && !strings.HasPrefix(item.OriginalURL, "https://") {
			return nil, errors.New("invalid URL format in batch item")
		}

		batchData[i] = storage.BatchURLData{
			CorrelationID: item.CorrelationID,
			OriginalURL:   item.OriginalURL,
			ShortURL:      generateID(),
		}
	}

	results, err := s.storage.SaveBatch(batchData, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to save batch URLs: %w", err)
	}

	resp := make([]BatchURLResponse, len(results))
	for i, result := range results {
		resp[i] = BatchURLResponse{
			CorrelationID: result.CorrelationID,
			ShortURL:      result.ShortURL,
		}
	}

	return resp, nil
}

func generateID() string {
	return uuid.New().String()
}
