package config

import (
	"context"
	"flag"

	"github.com/caarlos0/env"
)

type ContextOption func(*Context)

type Context struct {
	ctx             context.Context
	ServiceURL      string
	ServerAddress   string
	FileStoragePath string
}

func NewContext(opts ...ContextOption) *Context {
	uCtx := Context{}
	for _, o := range opts {
		o(&uCtx)
	}
	return &uCtx
}

func NewContextFromEnvAndCMD() *Context {
	var envronment Env
	err := env.Parse(&envronment)
	if err != nil {
		panic(err)
	}
	baseURL := flag.String("b", envronment.BaseURL, "a string")
	serverAddress := flag.String("s", envronment.ServerAddress, "a string")
	fileStoragePath := flag.String("f", envronment.FileStoragePath, "a string")
	flag.Parse()
	uCtx := NewContext(
		WithServiceURL(*baseURL),
		WithServerAddress(*serverAddress),
		WithFileStoragePath(*fileStoragePath),
	)
	return uCtx
}

func WithServerAddress(address string) ContextOption {
	return func(c *Context) {
		c.ServerAddress = address
	}
}

func WithServiceURL(serviceURL string) ContextOption {
	return func(c *Context) {
		c.ServiceURL = serviceURL
	}
}

func WithFileStoragePath(path string) ContextOption {
	return func(c *Context) {
		c.FileStoragePath = path
	}
}
