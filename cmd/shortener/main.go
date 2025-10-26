package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Create a context that will be canceled when a signal is received
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		cancel()
	}()

	if cfg.EnableHTTPS {
		startHTTPSServer(ctx, cfg, srv)
	} else {
		startHTTPServer(ctx, cfg, srv)
	}

	log.Println("Server gracefully stopped")
}

func startHTTPServer(ctx context.Context, cfg *config.Config, srv *server.Server) {
	log.Printf("Starting HTTP server on %s", cfg.ServerAddress)

	// Create the HTTP server
	httpServer := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: srv.Router,
	}

	// Start HTTP server in a goroutine
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for the context to be canceled
	<-ctx.Done()

	// Shutdown the server with a timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}

func startHTTPSServer(ctx context.Context, cfg *config.Config, srv *server.Server) {
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

	// Start HTTPS server in a goroutine
	go func() {
		if err := httpsServer.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start HTTPS server: %v", err)
		}
	}()

	// Wait for the context to be canceled
	<-ctx.Done()

	// Shutdown the server with a timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpsServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTPS server shutdown error: %v", err)
	}
}
