package cmd

import (
	"fmt"
	"github.com/zueve/go-shortener/internal/server"
)

func Run(server server.Server) {
	fmt.Println("Server started")
	server.Run()
}
