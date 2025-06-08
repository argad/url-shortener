package config

import (
	"flag"
	"fmt"
	"strings"
)

// Config структура для хранения конфигурации
type Config struct {
	ServerAddress string // Адрес запуска HTTP-сервера
	BaseShortURL  string // Базовый адрес результирующего сокращённого URL
}

// InitConfig функция для инициализации конфигурации с помощью флагов
func InitConfig() (*Config, error) {
	// Определяем флаги
	serverAddress := flag.String("a", "localhost:8080", "Адрес запуска HTTP-сервера (например, localhost:8888)")
	baseShortURL := flag.String("b", "http://localhost:8080", "Базовый адрес результирующего сокращённого URL (например, http://localhost:8000/qsd54gFg)")

	// Парсим флаги
	flag.Parse()

	// Проверяем корректность базового адреса URL
	if *baseShortURL == "" {
		return nil, fmt.Errorf("базовый адрес сокращённого URL не может быть пустым")
	}

	address := *serverAddress
	if strings.HasPrefix(address, "localhost:") {
		address = address[len("localhost:"):]
		address = ":" + address
	}

	// Создаём и возвращаем конфигурацию
	return &Config{
		ServerAddress: *serverAddress,
		BaseShortURL:  *baseShortURL,
	}, nil
}
