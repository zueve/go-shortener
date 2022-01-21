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
	ctx := config.NewContextFromEnvAndCMD()
	persistentStorage := storage.NewFileStorage(ctx.FileStoragePath)
	defer persistentStorage.Close()
	storageVar := storage.New(persistentStorage)
	serviceVar := services.New(storageVar)
	serverVar := server.New(ctx, serviceVar)
	go serverVar.ListenAndServe()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (kill -2)
	<-stop

	ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := serverVar.Shutdown(ctx2); err != nil {
		panic("unexpected err on graceful shutdown")
	}
	fmt.Println("main: done. exiting")
}
