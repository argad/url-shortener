package server

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/argad/url-shortener/internal/middleware"
	"github.com/argad/url-shortener/internal/storage"
)

// DeleteURLsHandler handles the deletion of URLs.
// It uses a channel and a worker pool to process deletion requests.
type DeleteURLsHandler struct {
	storage         storage.Storage
	deletionChannel chan DeletionTask
	workerPool      sync.WaitGroup
}

// DeletionTask represents a task for deleting a batch of URLs for a specific user.
type DeletionTask struct {
	URLs   []string
	UserID string
}

// NewDeleteURLsHandler creates a new DeleteURLsHandler and starts a pool of deletion workers.
func NewDeleteURLsHandler(storage storage.Storage) *DeleteURLsHandler {
	handler := &DeleteURLsHandler{
		storage:         storage,
		deletionChannel: make(chan DeletionTask, 100),
	}

	for i := 0; i < 5; i++ {
		go handler.deletionWorker()
	}

	return handler
}

// HandleDeleteURLs accepts a request to delete a batch of URLs for the authenticated user.
// It returns a 202 Accepted status on success.
func (h *DeleteURLsHandler) HandleDeleteURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var urls []string
	if err := json.NewDecoder(r.Body).Decode(&urls); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(urls) == 0 {
		http.Error(w, "No URLs to delete", http.StatusBadRequest)
		return
	}

	select {
	case h.deletionChannel <- DeletionTask{URLs: urls, UserID: userID}:
		w.WriteHeader(http.StatusAccepted)
	default:
		http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
	}
}

func (h *DeleteURLsHandler) deletionWorker() {
	for task := range h.deletionChannel {
		if err := h.storage.DeleteURLs(task.URLs, task.UserID); err != nil {
			println("Failed to delete URLs:", err.Error())
		}
	}
}

// Close closes the deletion channel.
func (h *DeleteURLsHandler) Close() {
	close(h.deletionChannel)
}
