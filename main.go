package main

import (
	"go-shortener/cmd/shortener"
	"go-shortener/internal/server"
	"go-shortener/internal/services"
	"go-shortener/internal/storage"
)

func main() {
	var storage_ = storage.New()
	var service_ = services.New(&storage_)
	var server_ = server.New(service_)

	cmd.Run(server_)
}
