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
// Thread Safety:
// All methods MUST be safe for concurrent use by multiple goroutines.
// Implementations are responsible for ensuring proper synchronization.
type Storage interface {
	// SaveURL stores a URL mapping.
	// Thread-safe: Yes
	SaveURL(url string, key string, userID string) (string, error)
	// GetURL retrieves the original URL by its short identifier.
	// Thread-safe: Yes
	GetURL(id string) (string, error)
	// SaveBatch stores multiple URL mappings in a single operation.
	// Thread-safe: Yes
	SaveBatch(batchData []BatchURLData, userID string) ([]BatchURLData, error)
	// GetUserURLs retrieves all URLs belonging to a specific user.
	// Thread-safe: Yes
	GetUserURLs(userID string) ([]URLData, error)
	// DeleteURLs marks multiple URLs as deleted.
	// Thread-safe: Yes
	DeleteURLs(shortURLs []string, userID string) error
}
