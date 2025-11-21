package config

import (
	"encoding/json"
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
	TrustedSubnet   string `env:"TRUSTED_SUBNET"`
}

// JSONConfig defines the structure of the JSON configuration file
type JSONConfig struct {
	ServerAddress   string `json:"server_address,omitempty"`
	BaseURL         string `json:"base_url,omitempty"`
	FileStoragePath string `json:"file_storage_path,omitempty"`
	DatabaseDSN     string `json:"database_dsn,omitempty"`
	EnableHTTPS     *bool  `json:"enable_https,omitempty"`
	TrustedSubnet   string `json:"trusted_subnet,omitempty"`
}

// InitConfig initializes the application's configuration.
// It reads configuration from environment variables and command-line flags,
// sets default values, and validates the configuration.
// Returns a populated Config struct or an error if initialization fails.
func InitConfig() (*Config, error) {
	cfg := setDefaults()

	// Get config file path from environment or flags
	configFilePath := getConfigFilePath()

	// Load JSON config first (lowest priority)
	if configFilePath != "" {
		if err := loadJSONConfig(cfg, configFilePath); err != nil {
			return nil, fmt.Errorf("error loading JSON config: %w", err)
		}
	}

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

func getConfigFilePath() string {
	// Check the command line flag first
	var configPath string
	flag.StringVar(&configPath, "c", "", "Path to JSON configuration file")
	flag.StringVar(&configPath, "config", "", "Path to JSON configuration file")

	// Parse only the config flag to get its value early
	for i, arg := range os.Args[1:] {
		if (arg == "-c" || arg == "-config") && i+1 < len(os.Args)-1 {
			return os.Args[i+2]
		}
		if strings.HasPrefix(arg, "-c=") {
			return strings.TrimPrefix(arg, "-c=")
		}
		if strings.HasPrefix(arg, "-config=") {
			return strings.TrimPrefix(arg, "-config=")
		}
	}

	// Check environment variable
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		return envConfig
	}

	return ""
}

func loadJSONConfig(cfg *Config, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	var jsonConfig JSONConfig
	if err := json.Unmarshal(data, &jsonConfig); err != nil {
		return fmt.Errorf("failed to parse JSON config: %w", err)
	}

	// Apply JSON config values only if they are not empty and current value is default
	if jsonConfig.ServerAddress != "" {
		cfg.ServerAddress = jsonConfig.ServerAddress
	}
	if jsonConfig.BaseURL != "" {
		cfg.BaseShortURL = jsonConfig.BaseURL
	}
	if jsonConfig.FileStoragePath != "" {
		cfg.FileStoragePath = jsonConfig.FileStoragePath
	}
	if jsonConfig.DatabaseDSN != "" {
		cfg.DatabaseDSN = jsonConfig.DatabaseDSN
	}
	if jsonConfig.EnableHTTPS != nil {
		cfg.EnableHTTPS = *jsonConfig.EnableHTTPS
	}
	if jsonConfig.TrustedSubnet != "" {
		cfg.TrustedSubnet = jsonConfig.TrustedSubnet
	}

	return nil
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
	trustedSubnet := flag.String("t", cfg.TrustedSubnet, "Trusted subnet in CIDR format")

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

	if !isEnvSet("TRUSTED_SUBNET") {
		cfg.TrustedSubnet = *trustedSubnet
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
		TrustedSubnet:   "",
	}
}

// isEnvSet checks if an environment variable is set
func isEnvSet(name string) bool {
	_, exists := os.LookupEnv(name)
	return exists
}
