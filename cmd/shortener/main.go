package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/zueve/go-shortener/internal/config"
	"github.com/zueve/go-shortener/internal/server"
	"github.com/zueve/go-shortener/internal/services"
	"github.com/zueve/go-shortener/internal/storage"
)

func main() {
	conf, err := config.NewFromEnvAndCMD()
	if err != nil {
		panic(err)
	}
	persistentStorage := storage.NewFileStorage(conf.FileStoragePath)
	defer persistentStorage.Close()
	storageVar := storage.New(persistentStorage)
	serviceVar := services.New(storageVar)
	serverVar := server.New(
		serviceVar,
		server.WithAddress(conf.ServerAddress),
		server.WithURL(conf.BaseURL),
	)
	go serverVar.ListenAndServe()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (kill -2)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := serverVar.Shutdown(ctx); err != nil {
		panic("unexpected err on graceful shutdown")
	}
	fmt.Println("main: done. exiting")
}
