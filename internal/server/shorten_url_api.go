package server

import (
	"encoding/json"
	"net/http"

	"github.com/argad/url-shortener/internal/middleware"
	"go.uber.org/zap"
)

// ShortenRequest defines the structure for a request to shorten a URL.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse defines the structure for the response of a URL shortening request.
type ShortenResponse struct {
	Result string `json:"result"`
}

// handleAPIShortenURL handles POST /api/shorten requests
func (s *Server) handleAPIShortenURL(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Use URLService instead of storage directly
	result, err := s.urlService.ShortenURL(userID, req.URL)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ShortenResponse{Result: result.ShortURL})
}
