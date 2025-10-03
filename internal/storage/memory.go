package storage

import (
	"fmt"
	"sync"
)

// InMemoryStorage is a storage implementation that uses an in-memory map to store URL data.
type InMemoryStorage struct {
	mu   sync.RWMutex
	data map[string]URLData
}

// NewInMemoryStorage creates a new InMemoryStorage.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]URLData),
	}
}

// SaveURL saves a new URL to the in-memory storage.
func (s *InMemoryStorage) SaveURL(url string, key string, userID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if url == "" {
		return "", fmt.Errorf("url cannot be empty")
	}

	//id := fmt.Sprintf("%d", len(s.data))
	s.data[key] = URLData{
		ShortURL:    key,
		OriginalURL: url,
		UserID:      userID,
		DeletedFlag: false,
	}

	return key, nil
}

// GetURL retrieves a URL from the in-memory storage.
func (s *InMemoryStorage) GetURL(id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if urlData, exists := s.data[id]; exists {

		if urlData.DeletedFlag {
			return "", &URLDeletedError{ShortURL: id}
		}

		return urlData.OriginalURL, nil
	}

	return "", fmt.Errorf("URL not found")

}

// GetUserURLs retrieves all URLs for a given user ID from the in-memory storage.
func (s *InMemoryStorage) GetUserURLs(userID string) ([]URLData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var userURLs []URLData

	for _, urlData := range s.data {
		if urlData.UserID == userID {
			userURLs = append(userURLs, urlData)
		}
	}

	return userURLs, nil
}

// SaveBatch saves a batch of URLs to the in-memory storage.
func (s *InMemoryStorage) SaveBatch(batchData []BatchURLData, userID string) ([]BatchURLData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var results []BatchURLData
	for _, item := range batchData {
		s.data[item.ShortURL] = URLData{
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

// DeleteURLs deletes a batch of URLs from the in-memory storage.
func (s *InMemoryStorage) DeleteURLs(shortURLs []string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, shortURL := range shortURLs {
		if urlData, exists := s.data[shortURL]; exists {
			if urlData.UserID == userID {
				urlData.DeletedFlag = true
				s.data[shortURL] = urlData
			}
		}
	}
	return nil
}
