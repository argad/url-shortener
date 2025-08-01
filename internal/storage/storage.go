package storage

import "fmt"

type BatchURLData struct {
	CorrelationID string
	OriginalURL   string
	ShortURL      string
	userID        string
}

type URLData struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id,omitempty"`
	DeletedFlag bool   `db:"is_deleted"`
}

type URLDeletedError struct {
	ShortURL string
}

func (e *URLDeletedError) Error() string {
	return fmt.Sprintf("URL %s has been deleted", e.ShortURL)
}

type Storage interface {
	SaveURL(url string, key string, userID string) (string, error)
	GetURL(id string) (string, error)
	SaveBatch(batchData []BatchURLData, userID string) ([]BatchURLData, error)
	GetUserURLs(userID string) ([]URLData, error)
	DeleteURLs(shortURLs []string, userID string) error
}
