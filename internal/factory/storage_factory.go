package factory

import (
	"fmt"
	"github.com/argad/url-shortener/internal/config"
	"github.com/argad/url-shortener/internal/database"
	storage2 "github.com/argad/url-shortener/internal/storage"
	"go.uber.org/zap"
)

// StorageFactory is a factory for creating storage instances.
type StorageFactory struct {
	logger *zap.Logger
}

// Logger is an interface for logging.
type Logger interface {
	Info(msg string)
	Error(msg string)
}

// StorageResult is the result of creating a storage instance.
// It contains the storage instance and a database connection if available.
type StorageResult struct {
	Storage storage2.Storage
	DB      *database.Database
}

// NewStorageFactory creates a new StorageFactory.
func NewStorageFactory(logger *zap.Logger) *StorageFactory {
	return &StorageFactory{logger: logger}
}

// CreateStorage creates a storage instance based on the provided configuration.
func (sf *StorageFactory) CreateStorage(cfg *config.Config) (*StorageResult, error) {
	storageType := sf.determineStorageType(cfg)

	switch storageType {
	case "postgres":
		return sf.createPostgresStorage(cfg.DatabaseDSN)
	case "file":
		return sf.createFileStorage(cfg.FileStoragePath)
	default:
		return sf.createInMemoryStorage()
	}
}

func (sf *StorageFactory) determineStorageType(cfg *config.Config) string {
	if cfg.DatabaseDSN != "" {
		return "postgres"
	}
	if cfg.FileStoragePath != "" {
		return "file"
	}
	return "memory"
}

func (sf *StorageFactory) createPostgresStorage(dsn string) (*StorageResult, error) {
	sf.logger.Info("Attempting to use PostgreSQL storage...")

	db, err := database.NewDatabase(dsn)
	if err != nil {
		sf.logger.Error(fmt.Sprintf("Failed to initialize PostgreSQL database: %v", err))
		return nil, fmt.Errorf("failed to initialize PostgreSQL database: %w", err)
	}

	postgresStorage, err := storage2.NewPostgresStorage(db.GetPool())
	if err != nil {
		sf.logger.Error(fmt.Sprintf("Failed to create PostgreSQL storage: %v", err))
		db.Close()
		return nil, fmt.Errorf("failed to create PostgreSQL storage: %w", err)
	}

	sf.logger.Info("Using PostgreSQL storage")
	return &StorageResult{
		Storage: postgresStorage,
		DB:      db,
	}, nil
}

func (sf *StorageFactory) createFileStorage(filePath string) (*StorageResult, error) {
	sf.logger.Info("Attempting to use file storage...")

	fileStorage, err := storage2.NewFileStorage(filePath)
	if err != nil {
		sf.logger.Error(fmt.Sprintf("Failed to create file storage: %v", err))
		return nil, fmt.Errorf("failed to create file storage: %w", err)
	}

	sf.logger.Info(fmt.Sprintf("Using file storage: %s", filePath))
	return &StorageResult{
		Storage: fileStorage,
		DB:      nil,
	}, nil
}

func (sf *StorageFactory) createInMemoryStorage() (*StorageResult, error) {
	sf.logger.Info("Using in-memory storage")

	memoryStorage := storage2.NewInMemoryStorage()
	return &StorageResult{
		Storage: memoryStorage,
		DB:      nil,
	}, nil
}
