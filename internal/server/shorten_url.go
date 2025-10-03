package server

import (
	"errors"
	"fmt"
	"github.com/argad/url-shortener/internal/middleware"
	"github.com/argad/url-shortener/internal/storage"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// handleShorten handles the creation of a new shortened URL.
// It reads the original URL from the request body, generates a unique ID,
// saves the URL to the storage, and returns the shortened URL in the response.
// If a URL conflict occurs, it returns the existing shortened URL with a conflict status.
func (s *Server) handleShorten(w http.ResponseWriter, r *http.Request) {
	originalURL, err := s.readAndValidateURL(r)
	userID, _ := middleware.GetUserID(r.Context())

	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id := generateID()
	urlKey, err := s.storage.SaveURL(originalURL, id, userID)
	if err != nil {
		s.handleSaveURLError(w, err)
		return
	}

	if err := s.writeShortURLResponse(w, urlKey, http.StatusCreated); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) readAndValidateURL(r *http.Request) (string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		return "", fmt.Errorf("invalid request body")
	}

	url := strings.TrimSpace(string(body))
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", fmt.Errorf("invalid URL format")
	}

	return url, nil
}

func (s *Server) handleSaveURLError(w http.ResponseWriter, err error) {
	var conflictErr *storage.URLConflictError
	if errors.As(err, &conflictErr) {
		w.WriteHeader(http.StatusConflict)
		if writeErr := s.writeShortURLResponse(w, conflictErr.ExistingShortURL, 0); writeErr != nil {
			s.logger.Error("Failed to write conflict response", zap.Error(writeErr))
		}
		return
	}

	s.logger.Error("Failed to save URL", zap.Error(err))
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (s *Server) writeShortURLResponse(w http.ResponseWriter, urlKey string, statusCode int) error {
	fullURL, err := url.JoinPath(s.baseURL, urlKey)
	if err != nil {
		return fmt.Errorf("failed to join URL path: %w", err)
	}

	w.Header().Set("Content-Type", "text/plain")
	if statusCode != 0 {
		w.WriteHeader(statusCode)
	}

	_, writeErr := w.Write([]byte(fullURL))
	return writeErr
}
