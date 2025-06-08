package storage

type Storage interface {
	SaveUrl(url string, key string) (string, error)
	GetUrl(id string) (string, error)
}
