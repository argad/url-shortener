package storage

type Storage interface {
	SaveURL(url string, key string) (string, error)
	GetURL(id string) (string, error)
}
