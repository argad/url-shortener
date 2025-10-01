package server

import (
	"fmt"
	"github.com/argad/url-shortener/internal/middleware"
	"github.com/argad/url-shortener/internal/storage"
	json "github.com/json-iterator/go"
	"go.uber.org/zap"
	"net/http"
	"net/url"
)

// handleAPIShortenBatch handles the batch creation of new shortened URLs from a JSON request.
// It decodes the JSON request containing a list of URLs, generates unique IDs for each,
// saves them to the storage in a batch, and returns a list of shortened URLs in a JSON response.
func (s *Server) handleAPIShortenBatch(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	if err := s.validateBatchRequestMethod(r); err != nil {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req, err := s.parseBatchRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	batchData, err := s.prepareBatchData(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results, err := s.storage.SaveBatch(batchData, userID)
	if err != nil {
		s.logger.Error("Failed to save batch URLs", zap.Error(err))
		http.Error(w, "Error saving URLs", http.StatusInternalServerError)
		return
	}

	if err := s.writeBatchResponse(w, results); err != nil {
		s.logger.Error("Failed to write batch response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) validateBatchRequestMethod(r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("method %s not allowed", r.Method)
	}
	return nil
}

func (s *Server) parseBatchRequest(r *http.Request) ([]BatchURLRequest, error) {
	var req []BatchURLRequest
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(&req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if len(req) == 0 {
		return nil, fmt.Errorf("empty batch")
	}

	return req, nil
}

func (s *Server) prepareBatchData(req []BatchURLRequest) ([]storage.BatchURLData, error) {
	batchData := make([]storage.BatchURLData, len(req))

	for i, item := range req {
		if err := s.validateBatchItem(item); err != nil {
			return nil, fmt.Errorf("invalid item at index %d: %w", i, err)
		}

		batchData[i] = storage.BatchURLData{
			CorrelationID: item.CorrelationID,
			OriginalURL:   item.OriginalURL,
			ShortURL:      generateID(),
		}
	}

	return batchData, nil
}

func (s *Server) validateBatchItem(item BatchURLRequest) error {
	if item.OriginalURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	if item.CorrelationID == "" {
		return fmt.Errorf("correlation ID cannot be empty")
	}

	return nil
}

func (s *Server) writeBatchResponse(w http.ResponseWriter, results []storage.BatchURLData) error {
	resp, err := s.buildBatchResponse(results)
	if err != nil {
		return fmt.Errorf("failed to build batch response: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return fmt.Errorf("failed to encode batch response: %w", err)
	}

	return nil
}

func (s *Server) buildBatchResponse(results []storage.BatchURLData) ([]BatchURLResponse, error) {
	resp := make([]BatchURLResponse, len(results))

	for i, result := range results {
		shortURL, err := url.JoinPath(s.baseURL, result.ShortURL)
		if err != nil {
			s.logger.Error("Failed to join URL path for batch response",
				zap.String("base_url", s.baseURL),
				zap.String("short_url", result.ShortURL),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to join URL path: %w", err)
		}

		resp[i] = BatchURLResponse{
			CorrelationID: result.CorrelationID,
			ShortURL:      shortURL,
		}
	}

	return resp, nil
}
