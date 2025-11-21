package server

import (
	"encoding/json"
	"net/http"

	"github.com/argad/url-shortener/internal/middleware"
	"go.uber.org/zap"
)

// handleDeleteUserURLs handles DELETE /api/user/urls requests
func (s *Server) handleDeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var shortCodes []string
	if err := json.NewDecoder(r.Body).Decode(&shortCodes); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(shortCodes) == 0 {
		http.Error(w, "Empty request", http.StatusBadRequest)
		return
	}

	// Use URLService
	err := s.urlService.DeleteURLs(userID, shortCodes)
	if err != nil {
		s.logger.Error("Failed to delete URLs", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
