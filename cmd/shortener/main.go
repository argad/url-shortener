package main

import (
	"github.com/argad/url-shortener/cmd/shortener/config"
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

	storageInstance := storage.NewInMemoryStorage()
	srv := server.NewServer(storageInstance, cfg.BaseShortURL)

	err2 := http.ListenAndServe(cfg.ServerAddress, srv.Router)
	if err2 != nil {
		panic(err2)
	}
}
