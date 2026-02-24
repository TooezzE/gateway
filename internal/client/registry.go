package client

import (
	"context"
	"errors"
)

type ClientsRegistry struct {
	clients map[string]Invoker
}

type Invoker interface {
	Invoke(ctx context.Context, method string, payload []byte) ([]byte, error)
}

func New(clients map[string]Invoker) *ClientsRegistry {
	return &ClientsRegistry{clients: clients}
}

func (r *ClientsRegistry) Get(service string) (Invoker, error) {
	c, ok := r.clients[service]
	if !ok {
		return nil, errors.New("unknown service")
	}
	return c, nil
}
