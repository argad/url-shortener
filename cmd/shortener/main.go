package main

import (
	"github.com/argad/url-shortener/cmd/shortener/server"
	"github.com/argad/url-shortener/cmd/shortener/storage"
	"net/http"
)

func main() {
	storageInstance := storage.NewInMemoryStorage()
	mux := server.NewServer(storageInstance)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
