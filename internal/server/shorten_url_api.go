package server

import (
	"errors"
	"fmt"
	"github.com/argad/url-shortener/internal/middleware"
	"github.com/argad/url-shortener/internal/storage"
	json "github.com/json-iterator/go"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strings"
)

// handleAPIShortenURL handles the creation of a new shortened URL from a JSON request.
// It decodes the JSON request, generates a unique ID, saves the URL to the storage,
// and returns the shortened URL in a JSON response.
// If a URL conflict occurs, it returns the existing shortened URL with a conflict status.
func (s *Server) handleAPIShortenURL(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	if err := s.validateJSONContentType(r); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	req, err := s.parseJSONRequest(r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id := generateID()
	urlKey, err := s.storage.SaveURL(req.URL, id, userID)
	if err != nil {
		s.handleJSONSaveURLError(w, req.URL, id, err)
		return
	}

	if err := s.writeJSONResponse(w, urlKey, http.StatusCreated); err != nil {
		s.logger.Error("Failed to write JSON response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) validateJSONContentType(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return fmt.Errorf("invalid content type: %s", contentType)
	}
	return nil
}

func (s *Server) parseJSONRequest(r *http.Request) (*ShortenRequest, error) {
	var req ShortenRequest
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(&req); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return &req, nil
}

func (s *Server) handleJSONSaveURLError(w http.ResponseWriter, originalURL, shortID string, err error) {
	s.logger.Error("Failed to save URL to storage",
		zap.String("original_url", originalURL),
		zap.String("short_id", shortID),
		zap.Error(err),
	)

	var conflictErr *storage.URLConflictError
	if errors.As(err, &conflictErr) {
		s.writeJSONConflictResponse(w, conflictErr.ExistingShortURL)
		return
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (s *Server) writeJSONConflictResponse(w http.ResponseWriter, existingShortURL string) {
	fullURL, err := url.JoinPath(s.baseURL, existingShortURL)
	if err != nil {
		s.logger.Error("Failed to join URL path for conflict response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict)

	response := map[string]string{"result": fullURL}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode conflict response", zap.Error(err))
	}
}

func (s *Server) writeJSONResponse(w http.ResponseWriter, urlKey string, statusCode int) error {
	shortURL, err := url.JoinPath(s.baseURL, urlKey)
	if err != nil {
		s.logger.Error("Failed to join short URL",
			zap.String("base_url", s.baseURL),
			zap.String("url_key", urlKey),
			zap.Error(err),
		)
		return fmt.Errorf("failed to join URL path: %w", err)
	}

	response := ShortenResponse{
		Result: shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return fmt.Errorf("failed to encode JSON response: %w", err)
	}

	return nil
}
