package storage

type BatchURLData struct {
	CorrelationID string
	OriginalURL   string
	ShortURL      string
}

type Storage interface {
	SaveURL(url string, key string) (string, error)
	GetURL(id string) (string, error)
	SaveBatch(batchData []BatchURLData) ([]BatchURLData, error)
}
