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
	errCh := make(chan error, 2)

	// старт gRPC LogGateway
	go func() {
		if err := StartLogGatewayServer(ctx, cfg.Admin.GRPCAddr, hub); err != nil {
			if ctx.Err() != nil {
				return
			}
			select {
			case errCh <- err:
			case <-ctx.Done():
			}
		}
	}()

	admin := NewAdminServer(cfg.Admin.Addr, reg, logger)
	admin.AddKernelControlHandlers()
	admin.AddLogStream(bus)

	// запустим сводку здоровья
	ha := NewHealthAggregator(reg, bus, logger)
	admin.SetHealthAggregator(ha)
	go ha.Run(ctx, 2*time.Second)

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
