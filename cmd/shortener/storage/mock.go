package storage

import "fmt"

type MockStorage struct {
	data map[string]URLData
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string]URLData),
	}
}

func (m *MockStorage) SaveURL(url string, key string, userID string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("url cannot be empty")
	}

	m.data[key] = URLData{
		ShortURL:    key,
		OriginalURL: url,
		UserID:      userID,
	}
	return key, nil
}

func (m *MockStorage) GetURL(id string) (string, error) {
	url, exists := m.data[id]
	if !exists {
		return "", fmt.Errorf("url with id %s not found", id)
	}
	return url.OriginalURL, nil
}

func (m *MockStorage) SaveBatch(batchData []BatchURLData, userID string) ([]BatchURLData, error) {
	results := make([]BatchURLData, len(batchData))
	for _, item := range batchData {
		m.data[item.ShortURL] = URLData{
			ShortURL:    item.ShortURL,
			OriginalURL: item.OriginalURL,
			UserID:      userID,
		}

		results = append(results, BatchURLData{
			CorrelationID: item.CorrelationID,
			OriginalURL:   item.OriginalURL,
			ShortURL:      item.ShortURL,
		})
	}

	return results, nil
}

func (m *MockStorage) GetUserURLs(userID string) ([]URLData, error) {

	var userURLs []URLData

	for _, urlData := range m.data {
		if urlData.UserID == userID {
			userURLs = append(userURLs, urlData)
		}
	}

	return userURLs, nil
}
