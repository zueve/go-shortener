package cmd

import (
	"fmt"
	"go-shortener/internal/server"
)

func Run(server server.Server) {
	fmt.Println("Server started")
	server.Run()
}
