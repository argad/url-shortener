package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/argad/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// handleGetURL handles the redirection of a shortened URL.
func (s *Server) handleGetURL(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")

	if shortURL == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Use URLService instead of storage directly
	url, isDeleted, err := s.urlService.GetOriginalURL(shortURL)
	if err != nil {
		if isDeleted {
			w.WriteHeader(http.StatusGone)
			return
		}

		// Check if it's a deleted URL error
		var deletedErr *storage.URLDeletedError
		if errors.As(err, &deletedErr) {
			w.WriteHeader(http.StatusGone)
			return
		}

		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// handlePing checks the health of the database connection.
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

// handleGetStats returns statistics about URLs and users
func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	// Use URLService
	urlCount, userCount, err := s.urlService.GetStats()
	if err != nil {
		s.logger.Error("Failed to get stats", zap.Error(err))
		http.Error(w, "Failed to get statistics", http.StatusInternalServerError)
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
		s.logger.Error("Failed to encode stats", zap.Error(err))
		http.Error(w, "Failed to encode stats", http.StatusInternalServerError)
	}
}
