package server

import (
	"encoding/json"
	"net/http"

	"github.com/argad/url-shortener/internal/middleware"
	"go.uber.org/zap"
)

// UserURLResponse defines the structure for a single URL returned to a user.
// It contains both the shortened URL and the original URL.
type UserURLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// handleGetUserURLs retrieves all URLs associated with the authenticated user.
// It fetches the URLs from the storage and returns them as a JSON array.
func (s *Server) handleGetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	urls, err := s.urlService.GetUserURLs(userID)
	if err != nil {
		s.logger.Error("Failed to get user URLs", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]UserURLResponse, len(urls))
	for i, url := range urls {
		response[i] = UserURLResponse{
			ShortURL:    s.baseURL + "/" + url.ShortURL,
			OriginalURL: url.OriginalURL,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
