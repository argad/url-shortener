package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/argad/url-shortener/cmd/shortener/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	host = "http://localhost:8080/"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type BatchURLRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchURLResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func generateID() string {

	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)[:8]
}

// POST create shortener /
func (s *Server) handleShorten(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	myUrl := strings.TrimSpace(string(body))
	if !strings.HasPrefix(myUrl, "http") {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id := generateID()
	urlKey, err := s.storage.SaveURL(myUrl, id)
	if err != nil {
		var conflictErr *storage.ErrURLConflict
		if errors.As(err, &conflictErr) {
			w.WriteHeader(http.StatusConflict)
			fullURL, err := url.JoinPath(s.baseURL, conflictErr.ExistingShortURL)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(fullURL))
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fullURL, err := url.JoinPath(s.baseURL, urlKey)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(fullURL))
}

// GET /{id}
func (s *Server) handleGetURL(w http.ResponseWriter, r *http.Request) {

	//test
	id := chi.URLParam(r, "id")

	if id == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	url, err := s.storage.GetURL(id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)

}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		s.logger.Error("Database connection is not initialized")
		http.Error(w, "Database connection is not available", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		s.logger.Error("Database ping failed", zap.Error(err))
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleAPIShortenURL(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var req ShortenRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	id := generateID()
	urlKey, err := s.storage.SaveURL(req.URL, id)
	if err != nil {
		//log
		s.logger.Error("Failed to save URL to storage",
			zap.String("original_url", req.URL),
			zap.String("short_id", id),
			zap.Error(err),
		)
		//
		var conflictErr *storage.ErrURLConflict
		if errors.As(err, &conflictErr) {

			fullURL, err := url.JoinPath(s.baseURL, conflictErr.ExistingShortURL)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"result": fullURL,
			})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shortURL, err := url.JoinPath(s.baseURL, urlKey)
	if err != nil {
		s.logger.Error("Failed to join short URL",
			zap.String("base_url", s.baseURL),
			zap.String("url_key", urlKey),
			zap.Error(err),
		)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := ShortenResponse{
		Result: shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	encoder.Encode(response)
}

func (s *Server) handleAPIShortenBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req []BatchURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req) == 0 {
		http.Error(w, "Empty batch", http.StatusBadRequest)
		return
	}

	batchData := make([]storage.BatchURLData, len(req))
	for i, item := range req {
		if item.OriginalURL == "" {
			http.Error(w, "URL cannot be empty", http.StatusBadRequest)
			return
		}

		id := generateID()
		batchData[i] = storage.BatchURLData{
			CorrelationID: item.CorrelationID,
			OriginalURL:   item.OriginalURL,
			ShortURL:      id,
		}
	}

	results, err := s.storage.SaveBatch(batchData)
	if err != nil {
		s.logger.Error(err.Error())
		http.Error(w, "Error saving URLs", http.StatusInternalServerError)
		return
	}

	resp := make([]BatchURLResponse, len(results))
	for i, result := range results {
		resp[i] = BatchURLResponse{
			CorrelationID: result.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", s.baseURL, result.ShortURL),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
