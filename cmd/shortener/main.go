package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/argad/url-shortener/internal/config"
	"github.com/argad/url-shortener/internal/factory"
	grpcserver "github.com/argad/url-shortener/internal/grpc"
	"github.com/argad/url-shortener/internal/grpc/pb"
	"github.com/argad/url-shortener/internal/server"
	"github.com/argad/url-shortener/internal/service"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("Ошибка инициализации конфигурации: %v", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger instance: %v", err)
	}
	defer logger.Sync()

	sf := factory.NewStorageFactory(logger)
	storageInstance, err := sf.CreateStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	if storageInstance.DB != nil {
		defer storageInstance.DB.Close()
	}

	srv, err := server.NewServer(storageInstance.Storage, cfg, storageInstance.DB)
	if err != nil {
		log.Fatalf("Failed to create new server: %v", err)
	}

	// Create URLService for gRPC
	urlService := service.NewURLService(storageInstance.Storage, cfg.BaseShortURL, logger)

	// Start HTTP server in goroutine
	httpServer := startHTTPServerAsync(cfg, srv, logger)

	// Start gRPC server if enabled
	var grpcSrv *grpc.Server
	if cfg.GRPCEnabled {
		grpcSrv = startGRPCServer(cfg, urlService, logger)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-sigChan

	log.Println("Shutting down servers...")

	// Shutdown HTTP server
	shutdownTimeout := 5 * time.Second
	if cfg.ShutdownTimeout > 0 {
		shutdownTimeout = cfg.ShutdownTimeout
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if httpServer != nil {
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP server shutdown error", zap.Error(err))
		}
	}

	// Shutdown gRPC server
	if grpcSrv != nil {
		grpcSrv.GracefulStop()
		logger.Info("gRPC server stopped")
	}

	log.Println("Servers gracefully stopped")
}

func startHTTPServerAsync(cfg *config.Config, srv *server.Server, logger *zap.Logger) *http.Server {
	var httpServer *http.Server

	if cfg.EnableHTTPS {
		httpServer = createHTTPSServer(cfg, srv, logger)
	} else {
		httpServer = &http.Server{
			Addr:         cfg.ServerAddress,
			Handler:      srv.Router,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		}
	}

	go func() {
		logger.Info("Starting HTTP server", zap.String("address", cfg.ServerAddress))
		var err error
		if cfg.EnableHTTPS {
			err = httpServer.ListenAndServeTLS("", "")
		} else {
			err = httpServer.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	return httpServer
}

func createHTTPSServer(cfg *config.Config, srv *server.Server, logger *zap.Logger) *http.Server {
	logger.Info("Configuring HTTPS with autocert",
		zap.String("domain", cfg.AutocertDomain))

	m := &autocert.Manager{
		Cache:      autocert.DirCache(cfg.AutocertDir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.AutocertDomain),
	}

	return &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      srv.Router,
		TLSConfig:    m.TLSConfig(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}

func startGRPCServer(cfg *config.Config, urlService *service.URLService, logger *zap.Logger) *grpc.Server {
	listener, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		logger.Fatal("Failed to listen for gRPC", zap.Error(err))
	}

	// Configure gRPC server options
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpcserver.NewAuthInterceptor(logger).Unary()),
	}

	// Add timeouts if configured
	if cfg.ReadTimeout > 0 {
		opts = append(opts, grpc.ConnectionTimeout(cfg.ReadTimeout))
	}

	grpcServer := grpc.NewServer(opts...)

	// Register service
	grpcService := grpcserver.NewServer(urlService, logger)
	pb.RegisterShortenerServiceServer(grpcServer, grpcService)

	go func() {
		logger.Info("Starting gRPC server", zap.String("address", cfg.GRPCAddress))
		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatal("gRPC server error", zap.Error(err))
		}
	}()

	return grpcServer
}
