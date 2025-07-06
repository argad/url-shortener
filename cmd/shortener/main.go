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

	db, err := database.NewDatabase(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	storageInstance, err := storage.NewFileStorage(cfg.EnvFilePath)
	if err != nil {
		log.Fatalf("Failed to create file storage: %v", err)
	}

	srv, err := server.NewServer(storageInstance, cfg.BaseShortURL, db)
	if err != nil {
		log.Fatalf("Failed to create new server: %v", err)
	}

	err2 := http.ListenAndServe(cfg.ServerAddress, srv.Router)
	if err2 != nil {
		panic(err2)
	}
}
