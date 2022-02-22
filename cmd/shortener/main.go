package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v4/stdlib"
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

	db, err := sqlx.Open("pgx", conf.DatabaseDSN)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err = storage.Migrate(db); err != nil {
		panic(err)
	}

	storageVar, err := storage.New(db, conf.DeleteSize, conf.DeleteWorkerCnt, conf.DeletePeriod)
	if err != nil {
		panic(err)
	}
	defer storageVar.Shutdown()

	serviceVar := services.New(storageVar)
	serverVar, err := server.New(
		serviceVar,
		server.WithAddress(conf.ServerAddress),
		server.WithURL(conf.BaseURL),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("Started at", conf.ServerAddress)
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
