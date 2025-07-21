package storage

import "fmt"

type InMemoryStorage struct {
	data map[string]URLData
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]URLData),
	}
}

func (s *InMemoryStorage) SaveURL(url string, key string, userID string) (string, error) {
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

func (s *InMemoryStorage) GetURL(id string) (string, error) {

	if urlData, exists := s.data[id]; exists {

		if urlData.DeletedFlag {
			return "", &URLDeletedError{ShortURL: id}
		}

		return urlData.OriginalURL, nil
	}

	return "", fmt.Errorf("URL not found")

}

func (s *InMemoryStorage) GetUserURLs(userID string) ([]URLData, error) {

	var userURLs []URLData

	for _, urlData := range s.data {
		if urlData.UserID == userID {
			userURLs = append(userURLs, urlData)
		}
	}

	return userURLs, nil
}

func (s *InMemoryStorage) SaveBatch(batchData []BatchURLData, userID string) ([]BatchURLData, error) {
	results := make([]BatchURLData, len(batchData))
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

func (s *InMemoryStorage) DeleteURLs(shortURLs []string, userID string) error {
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
