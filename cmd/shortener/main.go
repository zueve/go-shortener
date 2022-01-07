package main

import (
	"github.com/zueve/go-shortener/internal/server"
	"github.com/zueve/go-shortener/internal/services"
	"github.com/zueve/go-shortener/internal/storage"
)

func main() {
	var storageTest = storage.New()
	var serviceTest = services.New(&storageTest)
	var serverTest = server.New(serviceTest, "http://localhost:8080", 8080)

	serverTest.Run()
}
