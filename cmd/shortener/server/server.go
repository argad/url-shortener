package server

import (
	"github.com/argad/url-shortener/cmd/shortener/storage"
	"net/http"
)

type Server struct {
	storage storage.Storage
	mux     *http.ServeMux
}

func NewServer(storageInterface storage.Storage) *Server {

	s := &Server{
		storage: storageInterface,
		mux:     http.NewServeMux(),
	}

	s.mux.HandleFunc("/", s.HandleRequest)

	return s
}

// ServeHTTP делегирует обработку запроса встроенному mux
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
