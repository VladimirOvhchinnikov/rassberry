package main

import (
	"context"
	"fmt"

	"example.com/ffp/platform/ports"
)

type DomainKernelLauncher struct {
	reg    *DiscoveryRegistry
	bus    ports.EventBus
	logger ports.Logger
	rpc    ports.RPC
}

func NewDomainKernelLauncher(reg *DiscoveryRegistry, bus ports.EventBus, logger ports.Logger, rpc ports.RPC) *DomainKernelLauncher {
	return &DomainKernelLauncher{reg: reg, bus: bus, logger: logger, rpc: rpc}
}

func (l *DomainKernelLauncher) Launch(ctx context.Context, spec DomainSpec) error {
	return fmt.Errorf("launch mode %s not implemented", spec.Mode)
}
