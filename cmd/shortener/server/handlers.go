package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strings"
)

const (
	host = "http://localhost:8080/"
)

func generateID() string {

	//TODO: если ключ уже есть, то брать старый?
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)[:8]
}

// POST create shortener /
func (s *Server) handleShorten(w http.ResponseWriter, r *http.Request) {
	//contentType := r.Header.Get("Content-Type")
	//if !strings.HasPrefix(contentType, "text/plain") {
	//	http.Error(w, "Bad Request", http.StatusBadRequest)
	//	return
	//}

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
	urlKey, err := s.storage.SaveURL(url, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shortURL := s.baseURL + "/" + urlKey
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(shortURL))
}

// GET /{id}
func (s *Server) handleGetURL(w http.ResponseWriter, r *http.Request) {

	//test
	id := chi.URLParam(r, "id")

	if id == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	url, err := s.storage.GetURL(id)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)

}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

func (s *Server) handleAPIShortenURL(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var req ShortenRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	id := generateID()
	urlKey, err := s.storage.SaveURL(req.URL, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shortURL := s.baseURL + "/" + urlKey

	response := ShortenResponse{
		Result: shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	encoder.Encode(response)
}
