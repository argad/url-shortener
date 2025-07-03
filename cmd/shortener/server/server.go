package server

import (
	"fmt"
	"github.com/argad/url-shortener/cmd/shortener/middleware"
	"github.com/argad/url-shortener/cmd/shortener/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Server struct {
	storage storage.Storage
	Router  *chi.Mux
	baseURL string
	logger  *zap.Logger
}

func NewServer(storageInterface storage.Storage, baseURL string) (*Server, error) {

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	s := &Server{
		storage: storageInterface,
		Router:  chi.NewRouter(),
		baseURL: baseURL,
		logger:  logger,
	}

	s.Router.Use(middleware.LoggingMiddleware(s.logger))
	s.Router.Use(middleware.GzipMiddleware)

	s.routes()

	return s, nil
}

func (s *Server) routes() {
	s.Router.Post("/", s.handleShorten)
	s.Router.Get("/{id}", s.handleGetURL)
	s.Router.Post("/api/shorten", s.handleAPIShortenURL)

}
