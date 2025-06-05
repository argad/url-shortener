package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	host = "http://localhost:8080/"
)

var urlStore = make(map[string]string)

func generateID() string {

	//TODO: если ключ уже есть, то брать старый
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)[:8]
}

// POST create shortener /
func handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	url := strings.TrimSpace(string(body))
	if !strings.HasPrefix(url, "http") {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	id := generateID()
	urlStore[id] = url

	shortURL := host + id
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(shortURL))
}

// GET /{id}
func handleGetUrl(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet || r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	fmt.Println("Имя:", id)

	url, success := urlStore[id]

	if !success {
		http.Error(w, "Bad Request Not Found", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusTemporaryRedirect)
	_, _ = w.Write([]byte(url))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/" {
		handleShorten(w, r)
		return
	}

	if strings.HasPrefix(path, "/") && len(path) > 1 {
		variable := strings.TrimPrefix(path, "/")
		handleGetUrl(w, r, variable)
		return
	}

	// Если ничего не подошло, возвращаем 404
	http.Error(w, "Bad Request", http.StatusBadRequest)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRequest)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
