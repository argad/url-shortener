package server_test

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/argad/url-shortener/internal/auth"
	"github.com/argad/url-shortener/internal/config"
	"github.com/argad/url-shortener/internal/server"
	"github.com/argad/url-shortener/internal/storage"
)

func readGzipBody(resp *http.Response) (string, error) {
	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func ExampleServer_handleShorten() {
	// Create a new mock storage
	mockStorage := storage.NewMockStorage()

	// Create a new server with the mock storage
	cfg := &config.Config{BaseShortURL: "http://localhost:8080"}
	srv, err := server.NewServer(mockStorage, cfg, nil)
	if err != nil {
		fmt.Printf("Failed to create server: %v", err)
		return
	}

	// Create a new test server
	testServer := httptest.NewServer(srv.Router)
	defer testServer.Close()

	// Create a request to shorten a URL
	req, err := http.NewRequest("POST", testServer.URL+"/", strings.NewReader("https://google.com"))
	if err != nil {
		fmt.Printf("Failed to create request: %v", err)
		return
	}
	req.Header.Set("Accept-Encoding", "gzip")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := readGzipBody(resp)
	if err != nil {
		fmt.Printf("Failed to read response body: %v", err)
		return
	}

	// Print the status code and body
	fmt.Println(resp.StatusCode)
	fmt.Println(strings.Contains(body, "http://localhost:8080/"))

	// Output:
	// 201
	// true
}

func ExampleServer_handleAPIShortenURL() {
	// Create a new mock storage
	mockStorage := storage.NewMockStorage()

	// Create a new server with the mock storage
	cfg := &config.Config{BaseShortURL: "http://localhost:8080"}
	srv, err := server.NewServer(mockStorage, cfg, nil)
	if err != nil {
		fmt.Printf("Failed to create server: %v", err)
		return
	}

	// Create a new test server
	testServer := httptest.NewServer(srv.Router)
	defer testServer.Close()

	// Create a request to shorten a URL
	jsonBody := `{"url": "https://google.com"}`
	req, err := http.NewRequest("POST", testServer.URL+"/api/shorten", strings.NewReader(jsonBody))
	if err != nil {
		fmt.Printf("Failed to create request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := readGzipBody(resp)
	if err != nil {
		fmt.Printf("Failed to read response body: %v", err)
		return
	}

	// Print the status code and body
	fmt.Println(resp.StatusCode)
	fmt.Println(strings.Contains(body, `"result":"http://localhost:8080/`))

	// Output:
	// 201
	// true
}

func ExampleServer_handleGetURL() {
	// Create a new mock storage
	mockStorage := storage.NewMockStorage()
	key, _ := mockStorage.SaveURL("https://google.com", "test", "")

	// Create a new server with the mock storage
	cfg := &config.Config{BaseShortURL: "http://localhost:8080"}
	srv, err := server.NewServer(mockStorage, cfg, nil)
	if err != nil {
		fmt.Printf("Failed to create server: %v", err)
		return
	}

	// Create a new test server
	testServer := httptest.NewServer(srv.Router)
	defer testServer.Close()

	// Create a request to get a URL
	req, err := http.NewRequest("GET", testServer.URL+"/"+key, nil)
	if err != nil {
		fmt.Printf("Failed to create request: %v", err)
		return
	}

	// Create a client that does not follow redirects
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Print the status code and location header
	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get("Location"))

	// Output:
	// 307
	// https://google.com
}

func ExampleServer_handlePing() {
	// Create a new mock storage
	mockStorage := storage.NewMockStorage()

	// Create a new server with a nil database connection
	cfg := &config.Config{BaseShortURL: "http://localhost:8080"}
	srv, err := server.NewServer(mockStorage, cfg, nil)
	if err != nil {
		fmt.Printf("Failed to create server: %v", err)
		return
	}

	// Create a new test server
	testServer := httptest.NewServer(srv.Router)
	defer testServer.Close()

	// Create a request to ping the server
	req, err := http.NewRequest("GET", testServer.URL+"/ping", nil)
	if err != nil {
		fmt.Printf("Failed to create request: %v", err)
		return
	}

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Print the status code
	fmt.Println(resp.StatusCode)

	// Output:
	// 500
}

func ExampleServer_handleAPIShortenBatch() {
	// Create a new mock storage
	mockStorage := storage.NewMockStorage()

	// Create a new server with the mock storage
	cfg := &config.Config{BaseShortURL: "http://localhost:8080"}
	srv, err := server.NewServer(mockStorage, cfg, nil)
	if err != nil {
		fmt.Printf("Failed to create server: %v", err)
		return
	}

	// Create a new test server
	testServer := httptest.NewServer(srv.Router)
	defer testServer.Close()

	// Create a request to shorten a batch of URLs
	jsonBody := `[{"correlation_id": "1", "original_url": "https://google.com"}, {"correlation_id": "2", "original_url": "https://yandex.ru"}]`
	req, err := http.NewRequest("POST", testServer.URL+"/api/shorten/batch", strings.NewReader(jsonBody))
	if err != nil {
		fmt.Printf("Failed to create request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := readGzipBody(resp)
	if err != nil {
		fmt.Printf("Failed to read response body: %v", err)
		return
	}

	// Print the status code and body
	fmt.Println(resp.StatusCode)
	fmt.Println(strings.Contains(body, `"correlation_id":"1","short_url":"http://localhost:8080/`))
	fmt.Println(strings.Contains(body, `"correlation_id":"2","short_url":"http://localhost:8080/`))

	// Output:
	// 201
	// true
	// true
}

func ExampleServer_handleGetUserURLs() {
	// Create a new mock storage
	mockStorage := storage.NewMockStorage()
	_, _ = mockStorage.SaveURL("https://google.com", "test1", "user1")
	_, _ = mockStorage.SaveURL("https://yandex.ru", "test2", "user1")

	// Create a new server with the mock storage
	cfg := &config.Config{BaseShortURL: "http://localhost:8080"}
	srv, err := server.NewServer(mockStorage, cfg, nil)
	if err != nil {
		fmt.Printf("Failed to create server: %v", err)
		return
	}

	// Create a new test server
	testServer := httptest.NewServer(srv.Router)
	defer testServer.Close()

	// Create a request to get user URLs
	req, err := http.NewRequest("GET", testServer.URL+"/api/user/urls", nil)
	if err != nil {
		fmt.Printf("Failed to create request: %v", err)
		return
	}

	// Create a JWT token for the user
	token, err := auth.CreateJWTToken("user1")
	if err != nil {
		fmt.Printf("Failed to create JWT token: %v", err)
		return
	}

	// Add the cookie to the request
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
	req.Header.Set("Accept-Encoding", "gzip")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := readGzipBody(resp)
	if err != nil {
		fmt.Printf("Failed to read response body: %v", err)
		return
	}

	// Print the status code and body
	fmt.Println(resp.StatusCode)
	fmt.Println(strings.Contains(body, `"short_url":"http://localhost:8080/test1","original_url":"https://google.com"`))
	fmt.Println(strings.Contains(body, `"short_url":"http://localhost:8080/test2","original_url":"https://yandex.ru"`))

	// Output:
	// 200
	// true
	// true
}

func ExampleServer_handleDeleteUserURLs() {
	// Create a new mock storage
	mockStorage := storage.NewMockStorage()
	_, _ = mockStorage.SaveURL("https://google.com", "test1", "user1")

	// Create a new logger
	//logger, _ := zap.NewProduction()

	// Create a mock config
	cfg := &config.Config{
		BaseShortURL: "http://localhost:8080",
	}

	// Create a new server
	srv, err := server.NewServer(mockStorage, cfg, nil)
	if err != nil {
		fmt.Printf("Failed to create server: %v", err)
		return
	}

	// Create a new test server
	testServer := httptest.NewServer(srv.Router)
	defer testServer.Close()

	// Create a request to delete a URL
	jsonBody := `["test1"]`
	req, err := http.NewRequest("DELETE", testServer.URL+"/api/user/urls", strings.NewReader(jsonBody))
	if err != nil {
		fmt.Printf("Failed to create request: %v", err)
		return
	}

	// Create a JWT token for the user
	token, err := auth.CreateJWTToken("user1")
	if err != nil {
		fmt.Printf("Failed to create JWT token: %v", err)
		return
	}

	// Add the cookie to the request
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Print the status code
	fmt.Println(resp.StatusCode)

	// Output:
	// 202
}

func ExampleServer_handleShorten_badRequest() {
	// Create a new mock storage
	mockStorage := storage.NewMockStorage()

	// Create a new server with the mock storage
	cfg := &config.Config{BaseShortURL: "http://localhost:8080"}
	srv, err := server.NewServer(mockStorage, cfg, nil)
	if err != nil {
		fmt.Printf("Failed to create server: %v", err)
		return
	}

	// Create a new test server
	testServer := httptest.NewServer(srv.Router)
	defer testServer.Close()

	// Create a request with an empty body
	req, err := http.NewRequest("POST", testServer.URL+"/", strings.NewReader(""))
	if err != nil {
		fmt.Printf("Failed to create request: %v", err)
		return
	}

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Print the status code
	fmt.Println(resp.StatusCode)

	// Output:
	// 400
}

func ExampleServer_handleAPIShortenURL_invalidJSON() {
	// Create a new mock storage
	mockStorage := storage.NewMockStorage()

	// Create a new server with the mock storage
	cfg := &config.Config{BaseShortURL: "http://localhost:8080"}
	srv, err := server.NewServer(mockStorage, cfg, nil)
	if err != nil {
		fmt.Printf("Failed to create server: %v", err)
		return
	}

	// Create a new test server
	testServer := httptest.NewServer(srv.Router)
	defer testServer.Close()

	// Create a request with invalid JSON
	jsonBody := `{"url": "https://google.com"` // Invalid JSON
	req, err := http.NewRequest("POST", testServer.URL+"/api/shorten", strings.NewReader(jsonBody))
	if err != nil {
		fmt.Printf("Failed to create request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Print the status code
	fmt.Println(resp.StatusCode)

	// Output:
	// 400
}

func ExampleServer_handleGetURL_notFound() {
	// Create a new mock storage
	mockStorage := storage.NewMockStorage()

	// Create a new server with the mock storage
	cfg := &config.Config{BaseShortURL: "http://localhost:8080"}
	srv, err := server.NewServer(mockStorage, cfg, nil)
	if err != nil {
		fmt.Printf("Failed to create server: %v", err)
		return
	}

	// Create a new test server
	testServer := httptest.NewServer(srv.Router)
	defer testServer.Close()

	// Create a request to get a non-existent URL
	req, err := http.NewRequest("GET", testServer.URL+"/nonexistent", nil)
	if err != nil {
		fmt.Printf("Failed to create request: %v", err)
		return
	}

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Print the status code
	fmt.Println(resp.StatusCode)

	// Output:
	// 400
}
