package server

import (
	"fmt"

	"github.com/argad/url-shortener/internal/config"
	"github.com/argad/url-shortener/internal/database"
	middleware "github.com/argad/url-shortener/internal/middleware"
	"github.com/argad/url-shortener/internal/service"
	"github.com/argad/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Server represents the HTTP server for the URL shortener application.
// It holds all the dependencies required for the server to run, including
// the storage backend, router, base URL for shortened links, logger, and database connection.
type Server struct {
	storage    storage.Storage
	Router     *chi.Mux
	baseURL    string
	logger     *zap.Logger
	db         *database.Database
	config     *config.Config
	urlService *service.URLService
}

// NewServer creates and new Server instance.
func NewServer(storageInterface storage.Storage, cfg *config.Config, db *database.Database) (*Server, error) {

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	s := &Server{
		storage: storageInterface,
		Router:  chi.NewRouter(),
		baseURL: cfg.BaseShortURL,
		logger:  logger,
		db:      db,
		config:  cfg,
	}

	// Initialize URLService
	s.urlService = service.NewURLService(storageInterface, cfg.BaseShortURL, logger)

	s.Router.Use(middleware.LoggingMiddleware(s.logger))
	s.Router.Use(middleware.GzipMiddleware)
	s.Router.Use(middleware.AuthMiddleware)

	s.routes()

	return s, nil
}

func (s *Server) routes() {
	s.Router.Post("/", s.handleShorten)
	s.Router.Get("/ping", s.handlePing)
	s.Router.Post("/api/shorten", s.handleAPIShortenURL)
	s.Router.Get("/{id}", s.handleGetURL)
	s.Router.Post("/api/shorten/batch", s.handleAPIShortenBatch)
	s.Router.With(middleware.RequireAuth).Get("/api/user/urls", s.handleGetUserURLs)

	s.Router.With(middleware.RequireAuth).Delete("/api/user/urls", s.handleDeleteUserURLs)

	s.Router.With(middleware.CheckTrustedSubnet(s.config.TrustedSubnet)).Get("/api/internal/stats", s.handleGetStats)
}
