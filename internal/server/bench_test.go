package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/argad/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
)

func BenchmarkHandleShorten(b *testing.B) {
	s, err := NewServer(storage.NewInMemoryStorage(), "http://localhost:8080", nil)
	if err != nil {
		b.Fatalf("Failed to create new server: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString("https://example.com"))
		rr := httptest.NewRecorder()
		b.StartTimer()

		s.handleShorten(rr, req)

		if rr.Code != http.StatusCreated && rr.Code != http.StatusConflict {
			b.Errorf("unexpected status code: got %v, want %v or %v", rr.Code, http.StatusCreated, http.StatusConflict)
		}
	}
}

func BenchmarkHandleGetURL(b *testing.B) {
	memStorage := storage.NewInMemoryStorage()
	shortURL, err := memStorage.SaveURL("https://example.com", "test", "user1")
	if err != nil {
		b.Fatalf("Failed to save URL: %v", err)
	}

	s, err := NewServer(memStorage, "http://localhost:8080", nil)
	if err != nil {
		b.Fatalf("Failed to create new server: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		req := httptest.NewRequest("GET", "/"+shortURL, nil)
		rr := httptest.NewRecorder()

		// Add chi context for URL parameter
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", shortURL)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		b.StartTimer()

		s.handleGetURL(rr, req)

		if rr.Code != http.StatusTemporaryRedirect {
			b.Errorf("unexpected status code: got %v, want %v", rr.Code, http.StatusTemporaryRedirect)
		}
	}
}

func BenchmarkHandleAPIShortenURL(b *testing.B) {
	s, err := NewServer(storage.NewInMemoryStorage(), "http://localhost:8080", nil)
	if err != nil {
		b.Fatalf("Failed to create new server: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		payload := map[string]string{"url": "https://example.com"}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/shorten", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		b.StartTimer()

		s.handleAPIShortenURL(rr, req)

		if rr.Code != http.StatusCreated && rr.Code != http.StatusConflict {
			b.Errorf("unexpected status code: got %v, want %v or %v", rr.Code, http.StatusCreated, http.StatusConflict)
		}
	}
}
