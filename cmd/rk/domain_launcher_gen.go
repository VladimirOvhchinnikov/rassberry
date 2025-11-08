package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"example.com/ffp/platform/contracts"
	"example.com/ffp/platform/ports"
	rt "example.com/ffp/platform/runtime"
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
	switch spec.Mode {
	case "inproc":
		return l.launchInproc(ctx, spec)
	case "process":
		// каркас: регистрация placeholder + TODO
		l.logger.Log(ctx, "WARN", "process mode not implemented yet", map[string]any{"id": spec.ID})
		l.reg.Register(KernelRecord{
			ID:           spec.ID,
			Scope:        contracts.DomainScope,
			Manifest:     contracts.Manifest{KernelID: spec.ID, Version: "0.0.0", Scope: contracts.DomainScope, Features: []string{"process-placeholder"}},
			Health:       contracts.Health{Status: contracts.HealthDegraded, Since: time.Now(), Reason: "process mode placeholder"},
			RegisteredAt: time.Now(),
		})
		return nil
	case "remote":
		return errors.New("remote mode not implemented")
	default:
		return fmt.Errorf("unknown mode: %s", spec.Mode)
	}
}

func (l *DomainKernelLauncher) launchInproc(ctx context.Context, spec DomainSpec) error {
	// Хост домена
	host := rt.NewHost(spec.ID, contracts.DomainScope,
		rt.WithLogger(ports.NewTeeLogger(l.bus, spec.ID, string(contracts.DomainScope), spec.Kind)),
		rt.WithEventBus(l.bus),
		rt.WithRPC(l.rpc),
		rt.WithConfig(spec.Config),
	)

	// Фабрика домена "example" (встроенный демонстрационный)
	var kernel rt.KernelModule
	switch spec.Kind {
	case "example", "":
		kernel = newExampleDomain(spec.ID)
	default:
		return fmt.Errorf("unknown domain kind: %s", spec.Kind)
	}

	fsm := rt.NewFSM(kernel, host)
	if err := fsm.Run(ctx, spec.Config); err != nil {
		return err
	}

	// Регистрация в discovery
	rec := KernelRecord{
		ID:           spec.ID,
		Scope:        contracts.DomainScope,
		Manifest:     kernel.Manifest(),
		Exports:      &contracts.Exports{}, // пока пусто
		Health:       kernel.Health(),
		RegisteredAt: time.Now(),
	}
	l.reg.Register(rec)
	return nil
}

// ===== встроенный пример домена =====

type exampleDomain struct {
	rt.KernelModuleBase
	id     string
	host   rt.KernelHost
	health contracts.Health
	cancel context.CancelFunc
}

func newExampleDomain(id string) *exampleDomain {
	return &exampleDomain{id: id, health: contracts.Health{Status: contracts.HealthReady, Since: time.Now()}}
}

func (e *exampleDomain) Manifest() contracts.Manifest {
	return contracts.Manifest{KernelID: e.id, Version: "0.0.1", Scope: contracts.DomainScope, Features: []string{"http", "workers"}}
}

func (e *exampleDomain) OnLoad(ctx context.Context, host rt.KernelHost) error {
	e.host = host
	return nil
}

func (e *exampleDomain) OnStart(ctx context.Context) error {
	// простой heartbeat-воркер
	wctx, cancel := context.WithCancel(ctx)
	e.cancel = cancel
	go func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-wctx.Done():
				return
			case <-t.C:
				if e.host != nil && e.host.Logger() != nil {
					e.host.Logger().Log(ctx, "INFO", "example heartbeat", map[string]any{"worker": "hb"})
				}
			}
		}
	}()
	return nil
}

func (e *exampleDomain) OnStop(ctx context.Context) error {
	if e.cancel != nil {
		e.cancel()
	}
	return nil
}

func (e *exampleDomain) Health() contracts.Health { return e.health }
