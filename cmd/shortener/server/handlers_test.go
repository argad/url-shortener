package server

import (
	"bytes"
	"encoding/json"
	"github.com/argad/url-shortener/cmd/shortener/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			expectedBody:   "http://localhost:8080/", // Check only the prefix, as the ID is generated randomly
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			contentType:    "text/plain",
			body:           "http://example.com",
			expectedStatus: http.StatusMethodNotAllowed,
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
			mockStorage := storage.NewMockStorage()
			server, err := NewServer(mockStorage, "http://localhost:8080/", nil)
			if err != nil {
				t.Fatalf("Failed to create server: %v", err)
			}

			// Create test request
			req := httptest.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			// Create ResponseRecorder to record response
			rr := httptest.NewRecorder()

			// Call the handler being tested
			server.Router.ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Check response body
			if tt.expectedStatus == http.StatusCreated {
				// For a successful case, check only that the response starts with the expected host
				assert.True(t, strings.HasPrefix(rr.Body.String(), tt.expectedBody))
			}

			// Check Content-Type header for a successful case
			if tt.expectedStatus == http.StatusCreated {
				assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
			}
		})
	}
}

func TestServerHandleGetURL(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		urlID          string
		setupStorage   func(storage.Storage)
		expectedStatus int
		expectedURL    string // now checking the URL in the Location header
	}{
		{
			name:   "Successful URL retrieval",
			method: http.MethodGet,
			urlID:  "testid123",
			setupStorage: func(s storage.Storage) {
				s.SaveURL("http://example.com", "testid123")
			},
			expectedStatus: http.StatusTemporaryRedirect,
			expectedURL:    "http://example.com",
		},
		{
			name:           "Invalid method",
			method:         http.MethodPost,
			urlID:          "testid123",
			setupStorage:   func(s storage.Storage) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "URL not found",
			method:         http.MethodGet,
			urlID:          "nonexistent",
			setupStorage:   func(s storage.Storage) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock storage
			mockStorage := storage.NewMockStorage()

			// Configure storage for the test
			tt.setupStorage(mockStorage)

			server, err := NewServer(mockStorage, "http://localhost:8080/", nil)
			if err != nil {
				t.Fatalf("Failed to create server: %v", err)
			}

			// Create a test request
			req := httptest.NewRequest(tt.method, "/"+tt.urlID, nil)

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the handler being tested
			server.Router.ServeHTTP(rr, req)

			// Check the status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusTemporaryRedirect {
				// Check the Location header for the successful case
				assert.Equal(t, tt.expectedURL, rr.Header().Get("Location"))
			}
		})
	}
}

func TestServerHandleAPIShortenURL(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		contentType    string
		body           string
		expectedStatus int
		expectedResult bool // проверяем наличие поля result
	}{
		{
			name:           "Successful JSON URL shortening",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"url":"https://practicum.yandex.ru"}`,
			expectedStatus: http.StatusCreated,
			expectedResult: true,
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			contentType:    "application/json",
			body:           `{"url":"https://practicum.yandex.ru"}`,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedResult: false,
		},
		{
			name:           "Invalid Content-Type",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           `{"url":"https://practicum.yandex.ru"}`,
			expectedStatus: http.StatusBadRequest,
			expectedResult: false,
		},
		{
			name:           "Empty URL in JSON",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"url":""}`,
			expectedStatus: http.StatusInternalServerError,
			expectedResult: false,
		},
		{
			name:           "Invalid JSON format",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"url":}`,
			expectedStatus: http.StatusBadRequest,
			expectedResult: false,
		},
		{
			name:           "Missing url field",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"other":"https://practicum.yandex.ru"}`,
			expectedStatus: http.StatusInternalServerError,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := storage.NewMockStorage()
			server, err := NewServer(mockStorage, "http://localhost:8080/", nil)
			if err != nil {
				t.Fatalf("Failed to create server: %v", err)
			}

			req := httptest.NewRequest(tt.method, "/api/shorten", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()
			server.Router.ServeHTTP(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedResult {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

				var response ShortenResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.True(t, strings.HasPrefix(response.Result, "http://localhost:8080/"))
				assert.NotEmpty(t, strings.TrimPrefix(response.Result, "http://localhost:8080/"))
			}
		})
	}
}
