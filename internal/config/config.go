package config

import (
	"context"
)

type ContextOption func(*Context)

type Context struct {
	ctx        context.Context
	ServiceURL string
	Port       int
}

func NewContext(opts ...ContextOption) *Context {
	uCtx := Context{}
	for _, o := range opts {
		o(&uCtx)
	}
	return &uCtx
}

func WithPort(port int) ContextOption {
	return func(c *Context) {
		c.Port = port
	}
}

func WithServiceURL(serviceURL string) ContextOption {
	return func(c *Context) {
		c.ServiceURL = serviceURL
	}
}
