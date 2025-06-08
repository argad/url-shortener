package storage

import "fmt"

type InMemoryStorage struct {
	data map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]string),
	}
}

func (s *InMemoryStorage) SaveURL(url string, key string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("url cannot be empty")
	}

	//id := fmt.Sprintf("%d", len(s.data))
	s.data[key] = url
	return key, nil
}

func (s *InMemoryStorage) GetURL(id string) (string, error) {
	url, exists := s.data[id]
	if !exists {
		return "", fmt.Errorf("url with id %s not found", id)
	}
	return url, nil
}
