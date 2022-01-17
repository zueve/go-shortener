package config

import (
	"context"

	"github.com/caarlos0/env"
)

type ContextOption func(*Context)

type Context struct {
	ctx           context.Context
	ServiceURL    string
	ServerAddress string
}

func NewContext(opts ...ContextOption) *Context {
	uCtx := Context{}
	for _, o := range opts {
		o(&uCtx)
	}
	return &uCtx
}

func NewContextFormEnv() *Context {
	var envronment Env
	err := env.Parse(&envronment)
	if err != nil {
		panic(err)
	}
	uCtx := NewContext(
		WithServiceURL(envronment.BaseURL),
		WithServerAddress(envronment.ServerAddress),
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
