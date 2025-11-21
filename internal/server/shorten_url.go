package server

import (
	"io"
	"net/http"

	"github.com/argad/url-shortener/internal/middleware"
	"go.uber.org/zap"
)

// handleShorten handles POST / requests (plain text URL shortening)
func (s *Server) handleShorten(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := string(body)
	if originalURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Use URLService
	result, err := s.urlService.ShortenURL(userID, originalURL)
	if err != nil {
		s.logger.Error("Failed to shorten URL", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return 409 if URL already exists
	statusCode := http.StatusCreated
	if result.AlreadyExists {
		statusCode = http.StatusConflict
	}

	w.WriteHeader(statusCode)
	w.Write([]byte(result.ShortURL))
}
