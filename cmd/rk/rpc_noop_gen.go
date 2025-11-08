package main

import (
	"context"
	"sync"

	"example.com/ffp/platform/ports"
)

type noopRPC struct {
	mu       sync.Mutex
	services []any
}

var _ ports.RPC = (*noopRPC)(nil)

func (n *noopRPC) Register(s any) error {
	n.mu.Lock()
	n.services = append(n.services, s)
	n.mu.Unlock()
	return nil
}

func (n *noopRPC) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
