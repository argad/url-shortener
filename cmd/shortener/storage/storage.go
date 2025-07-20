package storage

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
}

type Storage interface {
	SaveURL(url string, key string, userID string) (string, error)
	GetURL(id string) (string, error)
	SaveBatch(batchData []BatchURLData, userID string) ([]BatchURLData, error)
	GetUserURLs(userID string) ([]URLData, error)
}
