package server

import (
	"encoding/json"
	"net/http"

	"github.com/argad/url-shortener/internal/middleware"
	"github.com/argad/url-shortener/internal/service"
	"go.uber.org/zap"
)

// BatchURLRequest defines the structure for a single URL in a batch shortening request.
type BatchURLRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchURLResponse defines the structure for a single URL in the response of a batch shortening request.
type BatchURLResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// handleAPIShortenBatch handles POST /api/shorten/batch requests
func (s *Server) handleAPIShortenBatch(w http.ResponseWriter, r *http.Request) {
	var batchReq []BatchURLRequest
	if err := json.NewDecoder(r.Body).Decode(&batchReq); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(batchReq) == 0 {
		http.Error(w, "Empty batch request", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert HTTP request to service format
	items := make([]service.BatchItem, len(batchReq))
	for i, req := range batchReq {
		items[i] = service.BatchItem{
			CorrelationID: req.CorrelationID,
			OriginalURL:   req.OriginalURL,
		}
	}

	// Use URLService
	results, err := s.urlService.ShortenURLBatch(userID, items)
	if err != nil {
		s.logger.Error("Failed to shorten batch", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert service results to HTTP response format
	response := make([]BatchURLResponse, len(results))
	for i, result := range results {
		response[i] = BatchURLResponse{
			CorrelationID: result.CorrelationID,
			ShortURL:      result.ShortURL,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
