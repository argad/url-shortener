package server

import (
	"github.com/argad/url-shortener/cmd/shortener/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	storage storage.Storage
	Router  *chi.Mux
	baseURL string
}

func NewServer(storageInterface storage.Storage, baseURL string) *Server {

	s := &Server{
		storage: storageInterface,
		Router:  chi.NewRouter(),
		baseURL: baseURL,
	}

	s.Router.Use(middleware.Logger)    //?
	s.Router.Use(middleware.Recoverer) //?

	s.routes()

	return s
}

func (s *Server) routes() {
	s.Router.Post("/", s.handleShorten)
	s.Router.Get("/{id}", s.handleGetURL)
}
