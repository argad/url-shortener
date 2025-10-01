package server

import (
	"fmt"

	"github.com/argad/url-shortener/internal/database"
	middleware "github.com/argad/url-shortener/internal/middleware"
	"github.com/argad/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Server represents the HTTP server for the URL shortener application.
// It holds all the dependencies required for the server to run, including
// the storage backend, router, base URL for shortened links, logger, and database connection.
type Server struct {
	storage storage.Storage
	Router  *chi.Mux
	baseURL string
	logger  *zap.Logger
	db      *database.Database
}

// NewServer creates and configures a new Server instance.
// It initializes the server with the given storage backend, base URL for shortened links,
// and a database connection. It also sets up the router with the necessary middleware and routes.
// Returns the newly created Server or an error if initialization fails.
func NewServer(storageInterface storage.Storage, baseURL string, db *database.Database) (*Server, error) {

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	s := &Server{
		storage: storageInterface,
		Router:  chi.NewRouter(),
		baseURL: baseURL,
		logger:  logger,
		db:      db,
	}

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

	s.Router.With(middleware.RequireAuth).Delete("/api/user/urls", NewDeleteURLsHandler(s.storage).HandleDeleteURLs)

}
