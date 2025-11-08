package main

import (
	"context"
	"fmt"
	"time"

	"example.com/ffp/platform/contracts"
	"example.com/ffp/platform/ports"
	rt "example.com/ffp/platform/runtime"
)

type DomainFactory func(id string) rt.KernelModule

var domainFactories = map[string]DomainFactory{}

// RegisterDomainFactory регистрирует фабрику домена по его kind.
func RegisterDomainFactory(kind string, f DomainFactory) {
	if kind == "" || f == nil {
		return
	}
	domainFactories[kind] = f
}

// LaunchDomainWithFactory пробует запустить домен через зарегистрированную фабрику.
// Возвращает handled=true, если фабрика найдена (даже при err != nil).
func LaunchDomainWithFactory(
	ctx context.Context,
	bus ports.EventBus,
	logger ports.Logger,
	rpc ports.RPC,
	reg *DiscoveryRegistry,
	spec DomainSpec,
) (handled bool, err error) {

	f, ok := domainFactories[spec.Kind]
	if !ok {
		return false, nil
	}

	host := rt.NewHost(spec.ID, contracts.DomainScope,
		ports.WithLogger(ports.NewTeeLogger(bus, spec.ID, string(contracts.DomainScope), spec.Kind)),
		ports.WithEventBus(bus),
		ports.WithRPC(rpc),
		ports.WithConfig(spec.Config),
	)

	kernel := f(spec.ID)
	fsm := rt.NewFSM(kernel, host)
	if err := fsm.Run(ctx, spec.Config); err != nil {
		return true, fmt.Errorf("run FSM: %w", err)
	}

	// Простейшая регистрация экспортов: HTTP `/hello`
	httpAddr := ":8081"
	if v, ok := spec.Config["http_addr"].(string); ok && v != "" {
		httpAddr = v
	}
	ex := contracts.Exports{
		Network: []contracts.NetworkEndpoint{{
			Name:      "hello",
			Protocol:  "http",
			Address:   httpAddr,
			Version:   "v1",
			Endpoints: []string{"/hello"},
		}},
	}

	rec := KernelRecord{
		ID:           spec.ID,
		Scope:        contracts.DomainScope,
		Manifest:     kernel.Manifest(),
		Exports:      &ex,
		Health:       kernel.Health(),
		RegisteredAt: time.Now(),
	}
	reg.Register(rec)
	return true, nil
}
