package storage

import "fmt"

type MockStorage struct {
	data map[string]string
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string]string),
	}
}

func (m *MockStorage) SaveURL(url string, key string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("url cannot be empty")
	}

	//id := fmt.Sprintf("%d", len(m.data))
	m.data[key] = url
	return key, nil
}

func (m *MockStorage) GetURL(id string) (string, error) {
	url, exists := m.data[id]
	if !exists {
		return "", fmt.Errorf("url with id %s not found", id)
	}
	return url, nil
}
