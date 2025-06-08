package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v10"
	"os"
	"strings"
)

// Config structure for storing configuration
type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseShortURL  string `env:"BASE_URL"`
}

// InitConfig function to initialize the configuration using flags
func InitConfig() (*Config, error) {
	var cfg Config

	cfg.ServerAddress = ":8080"
	cfg.BaseShortURL = "http://localhost:8080"

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error reading environment variables: %w", err)
	}

	serverAddress := flag.String("a", cfg.ServerAddress, "Address for starting the HTTP server (e.g., localhost:8888)")
	baseShortURL := flag.String("b", cfg.BaseShortURL, "Base address for the resulting shortened URL (e.g., http://localhost:8000/qsd54gFg)")

	// Parse the flags
	flag.Parse()

	if !isEnvSet("SERVER_ADDRESS") {
		cfg.ServerAddress = *serverAddress
	}

	if !isEnvSet("BASE_URL") {
		cfg.BaseShortURL = *baseShortURL
	}

	if cfg.BaseShortURL == "" {
		return nil, fmt.Errorf("the base address for the shortened URL cannot be empty")
	}

	// Нормализуем адрес сервера
	if cfg.ServerAddress != "" && !strings.Contains(cfg.ServerAddress, ":") {
		cfg.ServerAddress = ":" + cfg.ServerAddress
	}

	return &cfg, nil
}

// isEnvSet checks if an environment variable is set
func isEnvSet(name string) bool {
	_, exists := os.LookupEnv(name)
	return exists
}
