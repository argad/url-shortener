package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/argad/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ShortenRequest defines the structure for a request to shorten a URL.
// It contains the URL that needs to be shortened.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse defines the structure for the response of a URL shortening request.
// It contains the resulting shortened URL.
type ShortenResponse struct {
	Result string `json:"result"`
}

// BatchURLRequest defines the structure for a single URL in a batch shortening request.
// It includes a correlation ID to track the request and the original URL to be shortened.
type BatchURLRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchURLResponse defines the structure for a single URL in the response of a batch shortening request.
// It includes the correlation ID from the request and the resulting shortened URL.
type BatchURLResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func generateID() string {

	return uuid.New().String()
}

// handleGetURL handles the redirection of a shortened URL.
// It retrieves the original URL from storage based on the provided ID
// and redirects the client to it. If the URL is not found or has been deleted,
// it returns an appropriate HTTP error.
func (s *Server) handleGetURL(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")

	if shortURL == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	url, err := s.storage.GetURL(shortURL)
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

// handlePing checks the health of the database connection.
// It responds with HTTP 200 OK if the database is reachable,
// or HTTP 500 Internal Server Error if the connection is down.
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

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	urlCount, err := s.storage.GetURLCount()
	if err != nil {
		http.Error(w, "Failed to get URL count", http.StatusInternalServerError)
		return
	}

	userCount, err := s.storage.GetUserCount()
	if err != nil {
		http.Error(w, "Failed to get user count", http.StatusInternalServerError)
		return
	}

	stats := struct {
		URLs  int `json:"urls"`
		Users int `json:"users"`
	}{
		URLs:  urlCount,
		Users: userCount,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "Failed to encode stats", http.StatusInternalServerError)
	}
}
