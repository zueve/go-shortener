package main

import (
	"github.com/zueve/go-shortener/internal/config"
	"github.com/zueve/go-shortener/internal/server"
	"github.com/zueve/go-shortener/internal/services"
	"github.com/zueve/go-shortener/internal/storage"
)

func main() {
	storageVar := storage.New()
	serviceVar := services.New(storageVar)
	ctx := config.NewContext(
		config.WithServiceURL("http://localhost:8080"),
		config.WithPort(8080),
	)
	serverVar := server.New(ctx, serviceVar)

	serverVar.Run()
}
