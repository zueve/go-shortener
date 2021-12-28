package shortener

import (
	"fmt"
	"go-shortener/internal/app"
)

func Run() {
	fmt.Println("Starting...")
	server.RunServer()
}
