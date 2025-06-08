package server

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
)

const (
	host = "http://localhost:8080/"
)

func generateID() string {

	//TODO: если ключ уже есть, то брать старый
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)[:8]
}

// POST create shortener /
func (s *Server) handleShorten(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if r.Method != http.MethodPost || !strings.HasPrefix(contentType, "text/plain") {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	url := strings.TrimSpace(string(body))
	if !strings.HasPrefix(url, "http") {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	id := generateID()
	urlKey, err := s.storage.SaveUrl(url, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shortURL := host + urlKey
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(shortURL))
}

// GET /{id}
func (s *Server) handleGetUrl(w http.ResponseWriter, r *http.Request, id string) {
	contentType := r.Header.Get("Content-Type")
	if r.Method != http.MethodGet || !strings.HasPrefix(contentType, "text/plain") {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	url, success := s.storage.GetUrl(id)

	if success != nil {
		http.Error(w, "Bad Request Not Found", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusTemporaryRedirect)
	_, _ = w.Write([]byte(url))
}

func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/" {
		s.handleShorten(w, r)
		return
	}

	if strings.HasPrefix(path, "/") && len(path) > 1 {
		variable := strings.TrimPrefix(path, "/")
		s.handleGetUrl(w, r, variable)
		return
	}

	// Если ничего не подошло, возвращаем 404
	http.Error(w, "Bad Request", http.StatusBadRequest)
}
