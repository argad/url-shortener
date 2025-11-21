package storage

import "fmt"

// MockStorage is a mock storage implementation for testing.
type MockStorage struct {
	data map[string]URLData
}

// NewMockStorage creates a new MockStorage.
func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string]URLData),
	}
}

// SaveURL saves a new URL to the mock storage.
func (m *MockStorage) SaveURL(url string, key string, userID string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("url cannot be empty")
	}

	m.data[key] = URLData{
		ShortURL:    key,
		OriginalURL: url,
		UserID:      userID,
		DeletedFlag: false,
	}
	return key, nil
}

// GetURL retrieves a URL from the mock storage.
func (m *MockStorage) GetURL(id string) (string, error) {
	url, exists := m.data[id]
	if !exists {
		return "", fmt.Errorf("url with id %s not found", id)
	}
	if url.DeletedFlag {
		return "", &URLDeletedError{ShortURL: id}
	}

	return url.OriginalURL, nil
}

// SaveBatch saves a batch of URLs to the mock storage.
func (m *MockStorage) SaveBatch(batchData []BatchURLData, userID string) ([]BatchURLData, error) {
	var results []BatchURLData
	for _, item := range batchData {
		m.data[item.ShortURL] = URLData{
			ShortURL:    item.ShortURL,
			OriginalURL: item.OriginalURL,
			UserID:      userID,
			DeletedFlag: false,
		}

		results = append(results, BatchURLData{
			CorrelationID: item.CorrelationID,
			OriginalURL:   item.OriginalURL,
			ShortURL:      item.ShortURL,
		})
	}

	return results, nil
}

// GetUserURLs retrieves all URLs for a given user ID from the mock storage.
func (m *MockStorage) GetUserURLs(userID string) ([]URLData, error) {

	var userURLs []URLData

	for _, urlData := range m.data {
		if urlData.UserID == userID {
			userURLs = append(userURLs, urlData)
		}
	}

	return userURLs, nil
}

// DeleteURLs deletes a batch of URLs from the mock storage.
func (m *MockStorage) DeleteURLs(shortURLs []string, userID string) error {
	for _, shortURL := range shortURLs {
		if urlData, exists := m.data[shortURL]; exists {
			if urlData.UserID == userID {
				urlData.DeletedFlag = true
				m.data[shortURL] = urlData
			}
		}
	}
	return nil
}

// GetURLCount retrieves the total number of URLs.
func (m *MockStorage) GetURLCount() (int, error) {
	return len(m.data), nil
}

// GetUserCount retrieves the total number of users.
func (m *MockStorage) GetUserCount() (int, error) {
	users := make(map[string]struct{})
	for _, data := range m.data {
		if data.UserID != "" {
			users[data.UserID] = struct{}{}
		}
	}
	return len(users), nil
}
