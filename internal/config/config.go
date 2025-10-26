package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v10"
)

// Config defines the configuration parameters for the application.
// It includes settings for the server, database, storage, and JWT authentication.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseShortURL    string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	JWTSecret       string `env:"JWT_SECRET"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS"`
	AutocertDomain  string `env:"AUTOCERT_DOMAIN"`
	AutocertDir     string `env:"AUTOCERT_DIR"`
}

// InitConfig initializes the application's configuration.
// It reads configuration from environment variables and command-line flags,
// sets default values, and validates the configuration.
// Returns a populated Config struct or an error if initialization fails.
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
	jwtSecret := flag.String("j", cfg.JWTSecret, "JWT secret key for token signing")
	enableHTTPS := flag.Bool("s", cfg.EnableHTTPS, "Enable HTTPS server with Let's Encrypt certificates")
	autocertDomain := flag.String("domain", cfg.AutocertDomain, "Domain for Let's Encrypt certificate")
	autocertDir := flag.String("cert-dir", cfg.AutocertDir, "Directory to store Let's Encrypt certificates")

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

	if !isEnvSet("JWT_SECRET") {
		cfg.JWTSecret = *jwtSecret
	}
	if !isEnvSet("ENABLE_HTTPS") {
		cfg.EnableHTTPS = *enableHTTPS
	}

	if !isEnvSet("AUTOCERT_DOMAIN") {
		cfg.AutocertDomain = *autocertDomain
	}

	if !isEnvSet("AUTOCERT_DIR") {
		cfg.AutocertDir = *autocertDir
	}
}

func validate(cfg *Config) error {
	if cfg.BaseShortURL == "" {
		return fmt.Errorf("the base address for the shortened URL cannot be empty")
	}
	if cfg.JWTSecret == "" {
		return fmt.Errorf("JWT secret cannot be empty")
	}
	if cfg.EnableHTTPS {
		if cfg.AutocertDomain == "" {
			return fmt.Errorf("domain is required when HTTPS is enabled")
		}
		if cfg.AutocertDir == "" {
			return fmt.Errorf("certificate directory is required when HTTPS is enabled")
		}
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
		JWTSecret:       "test-phrase",
		EnableHTTPS:     false,
		AutocertDomain:  "",
		AutocertDir:     "./certs",
	}
}

// isEnvSet checks if an environment variable is set
func isEnvSet(name string) bool {
	_, exists := os.LookupEnv(name)
	return exists
}
