package main

import (
	"github.com/zueve/go-shortener/internal/server"
	"github.com/zueve/go-shortener/internal/services"
	"github.com/zueve/go-shortener/internal/storage"
)

func main() {
	storageTest := storage.New()
	serviceTest := services.New(storageTest)
	serverTest := server.New(serviceTest, "http://localhost:8080", 8080)

	serverTest.Run()
}
