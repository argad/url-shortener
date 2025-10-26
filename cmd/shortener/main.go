package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/argad/url-shortener/internal/config"
	"github.com/argad/url-shortener/internal/factory"
	"github.com/argad/url-shortener/internal/server"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
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

	if cfg.EnableHTTPS {
		startHTTPSServer(cfg, srv)
	} else {
		startHTTPServer(cfg, srv)
	}

	//err2 := http.ListenAndServe(cfg.ServerAddress, srv.Router)
	//if err2 != nil {
	//	panic(err2)
	//}
}

func startHTTPServer(cfg *config.Config, srv *server.Server) {
	log.Printf("Starting HTTP server on %s", cfg.ServerAddress)

	// Create the HTTP server
	httpServer := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: srv.Router,
	}

	// Start HTTP server
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}

}

func startHTTPSServer(cfg *config.Config, srv *server.Server) {
	log.Printf("Starting HTTPS server with autocert on %s for domain %s", cfg.ServerAddress, cfg.AutocertDomain)

	// Create autocert manager
	m := &autocert.Manager{
		Cache:      autocert.DirCache(cfg.AutocertDir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.AutocertDomain),
	}

	// Create the HTTPS server
	httpsServer := &http.Server{
		Addr:      cfg.ServerAddress,
		Handler:   srv.Router,
		TLSConfig: m.TLSConfig(),
	}

	// Start HTTPS server
	if err := httpsServer.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Failed to start HTTPS server: %v", err)
	}
}
