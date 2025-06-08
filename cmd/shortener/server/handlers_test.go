package server

import (
	"bytes"
	"github.com/argad/url-shortener/cmd/shortener/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServerHandleShorten(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		contentType    string
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Successful URL shortening",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           "http://example.com",
			expectedStatus: http.StatusCreated,
			expectedBody:   host, // Check only the prefix, as the ID is generated randomly
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			contentType:    "text/plain",
			body:           "http://example.com",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "Invalid Content-Type",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           "http://example.com",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "Empty URL",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "Invalid URL without http prefix",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           "example.com",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock storage
			mockStorage := storage.NewInMemoryStorage()
			server := &Server{
				storage: mockStorage,
				mux:     http.NewServeMux(),
			}

			// Create test request
			req := httptest.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			// Create ResponseRecorder to record response
			rr := httptest.NewRecorder()

			// Call the handler being tested
			server.handleShorten(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Check response body
			if tt.expectedStatus == http.StatusCreated {
				// For a successful case, check only that the response starts with the expected host
				assert.True(t, strings.HasPrefix(rr.Body.String(), tt.expectedBody))
			} else {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}

			// Check Content-Type header for a successful case
			if tt.expectedStatus == http.StatusCreated {
				assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
			}
		})
	}
}

func TestServerHandleGetUrl(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		contentType    string
		urlID          string
		setupStorage   func(storage.Storage)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Successful URL retrieval",
			method:      http.MethodGet,
			contentType: "text/plain",
			urlID:       "testid123",
			setupStorage: func(s storage.Storage) {
				s.SaveURL("http://example.com", "testid123")
			},
			expectedStatus: http.StatusTemporaryRedirect,
			expectedBody:   "http://example.com",
		},
		{
			name:           "Invalid method",
			method:         http.MethodPost,
			contentType:    "text/plain",
			urlID:          "testid123",
			setupStorage:   func(s storage.Storage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "Invalid Content-Type",
			method:         http.MethodGet,
			contentType:    "application/json",
			urlID:          "testid123",
			setupStorage:   func(s storage.Storage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "URL not found",
			method:         http.MethodGet,
			contentType:    "text/plain",
			urlID:          "nonexistent",
			setupStorage:   func(s storage.Storage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request Not Found\n",
		},
		{
			name:           "Empty ID",
			method:         http.MethodGet,
			contentType:    "text/plain",
			urlID:          "",
			setupStorage:   func(s storage.Storage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request Not Found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock storage
			mockStorage := storage.NewInMemoryStorage()

			// Set up storage for the test
			tt.setupStorage(mockStorage)

			server := &Server{
				storage: mockStorage,
				mux:     http.NewServeMux(),
			}

			// Create test request
			req := httptest.NewRequest(tt.method, "/"+tt.urlID, nil)
			req.Header.Set("Content-Type", tt.contentType)

			// Create ResponseRecorder to record response
			rr := httptest.NewRecorder()

			// Call the handler being tested
			server.handleGetUrl(rr, req, tt.urlID)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Check response body
			assert.Equal(t, tt.expectedBody, rr.Body.String())

			// Check Content-Type header for a successful case
			if tt.expectedStatus == http.StatusTemporaryRedirect {
				assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
			}
		})
	}
}
