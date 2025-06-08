package main

import (
	"github.com/argad/url-shortener/cmd/shortener/server"
	"github.com/argad/url-shortener/cmd/shortener/storage"
	"net/http"
)

func main() {
	storageInstance := storage.NewInMemoryStorage()
	srv := server.NewServer(storageInstance)

	err := http.ListenAndServe(`:8080`, srv.Router)
	if err != nil {
		panic(err)
	}
}
