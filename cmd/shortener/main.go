package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/argad/url-shortener/internal/config"
	"github.com/argad/url-shortener/internal/factory"
	"github.com/argad/url-shortener/internal/server"
	"go.uber.org/zap"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("Ошибка инициализации конфигурации: %v", err)
	}

	// TODO: put logger instance in config
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger instance: %v", err)
	}
	sf := factory.NewStorageFactory(logger)
	storageInstance, err := sf.CreateStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	// close db if exist
	if storageInstance.DB != nil {
		defer storageInstance.DB.Close()
	}

	srv, err := server.NewServer(storageInstance.Storage, cfg.BaseShortURL, storageInstance.DB)
	if err != nil {
		log.Fatalf("Failed to create new server: %v", err)
	}

	err2 := http.ListenAndServe(cfg.ServerAddress, srv.Router)
	if err2 != nil {
		panic(err2)
	}
}
