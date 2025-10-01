package storage

import "fmt"

// BatchURLData represents the data for a single URL in a batch operation.
type BatchURLData struct {
	CorrelationID string
	OriginalURL   string
	ShortURL      string
	userID        string
}

// URLData represents the data for a single URL.
type URLData struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id,omitempty"`
	DeletedFlag bool   `db:"is_deleted"`
}

// URLDeletedError is an error that occurs when a URL has been deleted.
type URLDeletedError struct {
	ShortURL string
}

// Error returns the error message.
func (e *URLDeletedError) Error() string {
	return fmt.Sprintf("URL %s has been deleted", e.ShortURL)
}

// Storage is the interface for URL storage.
type Storage interface {
	SaveURL(url string, key string, userID string) (string, error)
	GetURL(id string) (string, error)
	SaveBatch(batchData []BatchURLData, userID string) ([]BatchURLData, error)
	GetUserURLs(userID string) ([]URLData, error)
	DeleteURLs(shortURLs []string, userID string) error
}
