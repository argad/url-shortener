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
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseShortURL    string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

// InitConfig function to initialize the configuration using flags
func InitConfig() (*Config, error) {
	cfg := setDefaults()

	if err := parseEnvironment(cfg); err != nil {
		return nil, err
	}

	parseFlags(cfg)

	if err := validate(cfg); err != nil {
		return nil, err
	}

	normalizeServerAddress(cfg)

	return cfg, nil
}

func parseFlags(cfg *Config) {
	serverAddress := flag.String("a", cfg.ServerAddress, "Address for starting the HTTP server (e.g., localhost:8888)")
	baseShortURL := flag.String("b", cfg.BaseShortURL, "Base address for the resulting shortened URL (e.g., http://localhost:8000/qsd54gFg)")
	envFilePath := flag.String("f", cfg.FileStoragePath, "Base address of the file to storage")
	databaseDSN := flag.String("d", cfg.DatabaseDSN, "Base DSN address of the database")

	flag.Parse()

	if !isEnvSet("SERVER_ADDRESS") {
		cfg.ServerAddress = *serverAddress
	}

	if !isEnvSet("BASE_URL") {
		cfg.BaseShortURL = *baseShortURL
	}

	if !isEnvSet("FILE_STORAGE_PATH") {
		cfg.FileStoragePath = *envFilePath
	}

	if !isEnvSet("DATABASE_DSN") {
		cfg.DatabaseDSN = *databaseDSN
	}
}

func validate(cfg *Config) error {
	if cfg.BaseShortURL == "" {
		return fmt.Errorf("the base address for the shortened URL cannot be empty")
	}
	return nil
}

func normalizeServerAddress(cfg *Config) {
	if cfg.ServerAddress != "" && !strings.Contains(cfg.ServerAddress, ":") {
		cfg.ServerAddress = ":" + cfg.ServerAddress
	}
}

func parseEnvironment(cfg *Config) error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("error reading environment variables: %w", err)
	}
	return nil
}

func setDefaults() *Config {
	return &Config{
		ServerAddress:   ":8080",
		BaseShortURL:    "http://localhost:8080",
		FileStoragePath: "",
		DatabaseDSN:     "",
	}
}

// isEnvSet checks if an environment variable is set
func isEnvSet(name string) bool {
	_, exists := os.LookupEnv(name)
	return exists
}
