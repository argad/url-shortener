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
	ServerAddress string // Address for starting the HTTP server
	BaseShortURL  string // Base address for the resulting shortened URL
}

// InitConfig function to initialize the configuration using flags
func InitConfig() (*Config, error) {
	// Define flags
	cfg := Config{
		ServerAddress: "localhost:8080",
		BaseShortURL:  "http://localhost:8080",
	}

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error reading environment variables: %w", err)
	}

	serverAddress := flag.String("a", "localhost:8080", "Address for starting the HTTP server (e.g., localhost:8888)")
	baseShortURL := flag.String("b", "http://localhost:8080", "Base address for the resulting shortened URL (e.g., http://localhost:8000/qsd54gFg)")

	// Parse the flags
	flag.Parse()

	if !isEnvSet("SERVER_ADDRESS") && *serverAddress != cfg.ServerAddress {
		cfg.ServerAddress = *serverAddress
	}

	if !isEnvSet("BASE_URL") && *baseShortURL != cfg.BaseShortURL {
		cfg.BaseShortURL = *baseShortURL
	}

	if cfg.BaseShortURL == "" {
		return nil, fmt.Errorf("the base address for the shortened URL cannot be empty")
	}

	address := cfg.ServerAddress
	if strings.HasPrefix(address, "localhost:") {
		address = address[len("localhost:"):]
		address = ":" + address
	}

	cfg.ServerAddress = address

	return &cfg, nil
}

// isEnvSet checks if an environment variable is set
func isEnvSet(name string) bool {
	_, exists := os.LookupEnv(name)
	return exists
}
