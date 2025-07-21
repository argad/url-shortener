package server

import (
	"encoding/json"
	"github.com/argad/url-shortener/cmd/shortener/middleware"
	"net/http"
	"sync"
)

type DeleteURLsHandler struct {
	storage         Storage
	deletionChannel chan DeletionTask
	workerPool      sync.WaitGroup
}

type DeletionTask struct {
	URLs   []string
	UserID string
}

type Storage interface {
	DeleteURLs(shortURLs []string, userID string) error
}

func NewDeleteURLsHandler(storage Storage) *DeleteURLsHandler {
	handler := &DeleteURLsHandler{
		storage:         storage,
		deletionChannel: make(chan DeletionTask, 100), // Буфер для задач
	}

	for i := 0; i < 5; i++ {
		go handler.deletionWorker()
	}

	return handler
}

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

func (h *DeleteURLsHandler) Close() {
	close(h.deletionChannel)
}
