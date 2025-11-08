package main

import (
	"context"
	"time"

	"example.com/ffp/platform/contracts"
)

func RunRootKernel(ctx context.Context, cfg RootConfig) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	bus := NewInMemoryEventBus()
	logger := NewStdLogger("rk")

	reg := NewDiscoveryRegistry()
	reg.RegisterKernel(contracts.Manifest{KernelID: "rk", Scope: contracts.RootScope, Version: "0.0.1"})

	hub := NewLogHub(bus)

	// старт gRPC LogGateway
	go func() {
		_ = StartLogGatewayServer(ctx, cfg.Admin.GRPCAddr, hub)
	}()

	admin := NewAdminServer(cfg.Admin.Addr, reg, logger)
	admin.AddKernelControlHandlers()
	admin.AddLogStream(bus)

	// запустим сводку здоровья
	ha := NewHealthAggregator(reg, bus, logger)
	admin.SetHealthAggregator(ha)
	go ha.Run(ctx, 2*time.Second)

	errCh := make(chan error, 2)

	go func() {
		errCh <- admin.Start(ctx)
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}
