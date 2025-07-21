package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/argad/url-shortener/cmd/shortener/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"time"
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

		var URLDeletedError *storage.URLDeletedError
		if errors.As(err, &URLDeletedError) {
			w.WriteHeader(http.StatusGone) // 410
			return
		}

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
