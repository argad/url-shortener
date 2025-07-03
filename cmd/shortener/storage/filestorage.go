package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorage struct {
	data     map[string]string // shortURL -> originalURL
	filePath string
	mutex    sync.RWMutex
	counter  int
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		data:     make(map[string]string),
		filePath: filePath,
		counter:  0,
	}

	// Загружаем данные из файла при инициализации
	if err := fs.loadFromFile(); err != nil {
		return nil, err
	}

	return fs, nil
}

func (fs *FileStorage) SaveURL(originalURL, shortURL string) (string, error) {
	if originalURL == "" {
		return "", fmt.Errorf("url cannot be empty")
	}

	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.data[shortURL] = originalURL
	fs.counter++

	record := URLRecord{
		UUID:        fmt.Sprintf("%d", fs.counter),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	if err := fs.appendToFile(record); err != nil {
		return "", err
	}

	return shortURL, nil
}

func (fs *FileStorage) GetURL(shortURL string) (string, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	url, exists := fs.data[shortURL]
	if !exists {
		return "", fmt.Errorf("url with id %s not found", shortURL)
	}
	return url, nil
}

func (fs *FileStorage) loadFromFile() error {
	file, err := os.OpenFile(fs.filePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	maxCounter := 0

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var record URLRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			continue
		}

		fs.data[record.ShortURL] = record.OriginalURL

		var uuid int
		if _, err := fmt.Sscanf(record.UUID, "%d", &uuid); err == nil {
			if uuid > maxCounter {
				maxCounter = uuid
			}
		}
	}

	fs.counter = maxCounter
	return scanner.Err()
}

func (fs *FileStorage) appendToFile(record URLRecord) error {
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
