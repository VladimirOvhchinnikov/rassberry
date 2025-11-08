package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"example.com/ffp/platform/contracts"
	"example.com/ffp/platform/ports"
	rt "example.com/ffp/platform/runtime"
)

// RunRootKernel настраивает DI (bus, logger, rpc), запускает админку и домены.
func RunRootKernel(ctx context.Context, configPath string) error {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Шина событий (in-proc)
	bus := rt.NewInprocEventBus(cfg.Telemetry.Buffer)

	// Локальный stdout sink + tee в EventBus
	var logger ports.Logger = ports.NewTeeLogger(bus, "rk", string(contracts.RootScope), "admin")

	// RPC-заглушка
	var rpc ports.RPC = &noopRPC{}

	// Регистрация/дискавери
	reg := NewDiscoveryRegistry()

	// Admin HTTP
	admin := NewAdminServer(cfg.Admin.Addr, reg)
	go func() {
		// Admin живёт в своём контексте, пока не отменят общий
		_ = admin.Start(ctx)
	}()

	// Запуск доменов
	launcher := NewDomainKernelLauncher(reg, bus, logger, rpc)

	for _, d := range cfg.Domains {
		if err := launcher.Launch(ctx, d); err != nil {
			logger.Log(ctx, "ERROR", "failed to launch domain", map[string]any{"id": d.ID, "err": err.Error()})
			continue
		}
		logger.Log(ctx, "INFO", "domain launched", map[string]any{"id": d.ID, "mode": d.Mode, "kind": d.Kind})
	}

	// Ожидание сигналов
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
	case <-sigCh:
	}
	return nil
}
