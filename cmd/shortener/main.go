package main

import (
	"github.com/zueve/go-shortener/internal/server"
	"github.com/zueve/go-shortener/internal/services"
	"github.com/zueve/go-shortener/internal/storage"
)

func main() {
	var storage_ = storage.New()
	var service_ = services.New(&storage_)
	var server_ = server.New(service_, "http://localhost:8080", 8080)

	server_.Run()
}
