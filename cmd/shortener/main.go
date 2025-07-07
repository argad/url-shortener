package main

import (
	"github.com/argad/url-shortener/cmd/shortener/config"
	"github.com/argad/url-shortener/cmd/shortener/database"
	"github.com/argad/url-shortener/cmd/shortener/server"
	"github.com/argad/url-shortener/cmd/shortener/storage"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("Ошибка инициализации конфигурации: %v", err)
	}

	storageInstance, db, err := createStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	// close db if exist
	if db != nil {
		defer db.Close()
	}

	//db, err := database.NewDatabase(cfg.DatabaseDSN)
	//if err != nil {
	//	log.Fatalf("Failed to initialize database: %v", err)
	//}
	//defer db.Close()

	//storageInstance, err = storage.NewFileStorage(cfg.EnvFilePath)
	//if err != nil {
	//	log.Fatalf("Failed to create file storage: %v", err)
	//}

	srv, err := server.NewServer(storageInstance, cfg.BaseShortURL, db)
	if err != nil {
		log.Fatalf("Failed to create new server: %v", err)
	}

	err2 := http.ListenAndServe(cfg.ServerAddress, srv.Router)
	if err2 != nil {
		panic(err2)
	}
}

func createStorage(cfg *config.Config) (storage.Storage, *database.Database, error) {

	if cfg.DatabaseDSN != "" {
		log.Println("Attempting to use PostgreSQL storage...")
		db, err := database.NewDatabase(cfg.DatabaseDSN)
		if err != nil {
			log.Printf("Failed to initialize PostgreSQL database: %v", err)
		} else {
			postgresStorage, err := storage.NewPostgresStorage(db.GetPool())
			if err != nil {
				log.Printf("Failed to create PostgreSQL storage: %v", err)
				db.Close()
			} else {
				log.Println("Using PostgreSQL storage")
				return postgresStorage, db, nil
			}
		}
	}

	if cfg.EnvFilePath != "" {
		log.Println("Attempting to use file storage...")
		fileStorage, err := storage.NewFileStorage(cfg.EnvFilePath)
		if err != nil {
			log.Printf("Failed to create file storage: %v", err)
		} else {
			log.Printf("Using file storage: %s", cfg.EnvFilePath)
			return fileStorage, nil, nil
		}
	}

	log.Println("Using in-memory storage")
	memoryStorage := storage.NewInMemoryStorage()
	return memoryStorage, nil, nil
}
