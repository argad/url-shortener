package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type FileStorage struct {
	data     map[string]URLData // shortURL -> originalURL
	filePath string
	mutex    sync.Mutex
	counter  int
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		data:     make(map[string]URLData),
		filePath: filePath,
		counter:  0,
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directories for file path: %w", err)
	}

	if err := fs.loadFromFile(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return fs, nil
}

func (fs *FileStorage) SaveURL(originalURL, shortURL string, userID string) (string, error) {
	if originalURL == "" {
		return "", fmt.Errorf("url cannot be empty")
	}

	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	record := URLData{
		UserID:      userID,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	if err := fs.appendToFile(record); err != nil {
		return "", err
	}

	return shortURL, nil
}

func (fs *FileStorage) GetURL(shortURL string) (string, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	url, exists := fs.data[shortURL]
	if !exists {
		return "", fmt.Errorf("url with id %s not found", shortURL)
	}
	return url.OriginalURL, nil
}

func (fs *FileStorage) GetUserURLs(userID string) ([]URLData, error) {
	var userURLs []URLData

	for _, urlData := range fs.data {
		if urlData.UserID == userID {
			userURLs = append(userURLs, urlData)
		}
	}

	return userURLs, nil
}

func (fs *FileStorage) SaveBatch(batchData []BatchURLData, userID string) ([]BatchURLData, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	results := make([]BatchURLData, len(batchData))
	records := make([]URLData, len(batchData))

	for i, item := range batchData {

		records[i] = URLData{
			UserID:      userID,
			ShortURL:    item.ShortURL,
			OriginalURL: item.OriginalURL,
		}
		results[i] = item
	}

	if err := fs.appendBatchToFile(records); err != nil {
		return nil, err
	}

	return results, nil
}

func (fs *FileStorage) loadFromFile() error {
	file, err := os.OpenFile(fs.filePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var record URLData
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			continue
		}

		fs.data[record.ShortURL] = URLData{
			ShortURL:    record.ShortURL,
			OriginalURL: record.OriginalURL,
			UserID:      record.UserID,
		}

	}

	return scanner.Err()
}

func (fs *FileStorage) appendToFile(record URLData) error {
	file, err := os.OpenFile(fs.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	_, err = file.Write(append(data, '\n'))
	return err
}

func (fs *FileStorage) appendBatchToFile(records []URLData) error {
	file, err := os.OpenFile(fs.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, record := range records {
		data, err := json.Marshal(record)
		if err != nil {
			return err
		}

		if _, err := file.Write(append(data, '\n')); err != nil {
			return err
		}
	}

	return nil
}
