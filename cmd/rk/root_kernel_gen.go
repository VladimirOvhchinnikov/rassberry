package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/ffp/platform/contracts"
	"example.com/ffp/platform/ports"
)

var configPath string

func RunRootKernel(ctx context.Context, cfg RootConfig) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	bus := NewInMemoryEventBus()
	logger := NewStdLogger("rk")

	reg := NewDiscoveryRegistry()
	reg.RegisterKernel(contracts.Manifest{KernelID: "rk", Scope: contracts.RootScope, Version: "0.0.1"})

	dp := NewDegradationPolicy(reg)
	go dp.Run(ctx, time.Second)

	hub := NewLogHub(bus)

	// старт gRPC LogGateway
	go func() {
		_ = StartLogGatewayServer(ctx, cfg.Admin.GRPCAddr, hub)
	}()

	admin := NewAdminServer(cfg.Admin.Addr, reg, logger)
	admin.AddKernelControlHandlers()
	admin.AddLogStream(bus)
	admin.AddTelemetryHandlers()

	// запустим сводку здоровья
	ha := NewHealthAggregator(reg, bus, logger)
	admin.SetHealthAggregator(ha)
	go ha.Run(ctx, 2*time.Second)

	var rpc ports.RPC

	launcher := NewDomainKernelLauncher(reg, bus, logger, rpc)
	mgr := NewDomainManager(reg, bus, logger, rpc)

	for _, d := range cfg.Domains {
		if d.Mode == "inproc" {
			if handled, err := func() (bool, error) {
				if _, ok := domainFactories[d.Kind]; ok {
					if err := mgr.launchInproc(ctx, d); err != nil {
						return true, err
					}
					return true, nil
				}
				return false, nil
			}(); handled {
				if err != nil {
					logger.Log(ctx, "ERROR", "manager launch failed", map[string]any{"id": d.ID, "kind": d.Kind, "err": err.Error()})
				} else {
					logger.Log(ctx, "INFO", "domain launched via manager", map[string]any{"id": d.ID, "kind": d.Kind})
				}
				continue
			}
		}

		// Сначала пробуем через фабрику (site и др.)
		if handled, err := LaunchDomainWithFactory(ctx, bus, logger, rpc, reg, d); handled {
			if err != nil {
				logger.Log(ctx, "ERROR", "factory launch failed", map[string]any{"id": d.ID, "kind": d.Kind, "err": err.Error()})
			} else {
				logger.Log(ctx, "INFO", "domain launched via factory", map[string]any{"id": d.ID, "mode": d.Mode, "kind": d.Kind})
			}
			continue // не вызываем старый путь
		}

		if err := launcher.Launch(ctx, d); err != nil {
			logger.Log(ctx, "ERROR", "domain launch failed", map[string]any{"id": d.ID, "mode": d.Mode, "kind": d.Kind, "err": err.Error()})
		} else {
			logger.Log(ctx, "INFO", "domain launch scheduled", map[string]any{"id": d.ID, "mode": d.Mode, "kind": d.Kind})
		}
	}

	errCh := make(chan error, 2)

	go func() {
		errCh <- admin.Start(ctx)
	}()

	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	go func() {
		for range hup {
			if cfg2, err := LoadConfig(configPath); err == nil {
				mgr.Reload(ctx, cfg2.Domains)
				logger.Log(ctx, "INFO", "config reloaded (domains)", map[string]any{"count": len(cfg2.Domains)})
			} else {
				logger.Log(ctx, "ERROR", "config reload failed", map[string]any{"err": err.Error()})
			}
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}
