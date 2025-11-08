package main

import (
	"context"
	"reflect"
	"time"

	"example.com/ffp/platform/contracts"
	"example.com/ffp/platform/ports"
	rt "example.com/ffp/platform/runtime"
)

type domainRun struct {
	spec   DomainSpec
	cancel context.CancelFunc
	fsm    *rt.FSM
	kernel rt.KernelModule
}

type DomainManager struct {
	reg    *DiscoveryRegistry
	bus    ports.EventBus
	logger ports.Logger
	rpc    ports.RPC

	runs map[string]*domainRun
}

func NewDomainManager(reg *DiscoveryRegistry, bus ports.EventBus, logger ports.Logger, rpc ports.RPC) *DomainManager {
	return &DomainManager{reg: reg, bus: bus, logger: logger, rpc: rpc, runs: make(map[string]*domainRun)}
}

func (m *DomainManager) manages(spec DomainSpec) bool {
	mode := spec.Mode
	if mode == "" {
		mode = "inproc"
	}
	if mode != "inproc" {
		return false
	}
	_, ok := domainFactories[spec.Kind]
	return ok
}

func (m *DomainManager) launchInproc(ctx context.Context, spec DomainSpec) error {
	f, ok := domainFactories[spec.Kind]
	if !ok {
		// нет фабрики — пусть старый лаунчер решает
		return NewDomainKernelLauncher(m.reg, m.bus, m.logger, m.rpc).Launch(ctx, spec)
	}
	host := rt.NewHost(spec.ID, contracts.DomainScope,
		ports.WithLogger(ports.NewTeeLogger(m.bus, spec.ID, string(contracts.DomainScope), spec.Kind)),
		ports.WithEventBus(m.bus),
		ports.WithRPC(m.rpc),
		ports.WithConfig(spec.Config),
	)
	k := f(spec.ID)
	fsm := rt.NewFSM(k, host)

	dctx, cancel := context.WithCancel(ctx)
	if err := fsm.Run(dctx, spec.Config); err != nil {
		cancel()
		return err
	}

	// exports: HTTP /hello если задано http_addr
	httpAddr := ":8081"
	if v, ok := spec.Config["http_addr"].(string); ok && v != "" {
		httpAddr = v
	}
	ex := &contracts.Exports{
		Network: []contracts.NetworkEndpoint{{Name: "hello", Protocol: "http", Address: httpAddr, Version: "v1", Endpoints: []string{"/hello"}}},
	}
	m.reg.Register(KernelRecord{
		ID: spec.ID, Scope: contracts.DomainScope, Manifest: k.Manifest(), Exports: ex,
		Health: k.Health(), RegisteredAt: time.Now(),
	})

	m.runs[spec.ID] = &domainRun{spec: spec, cancel: cancel, fsm: fsm, kernel: k}
	return nil
}

func (m *DomainManager) stop(id string) {
	if r := m.runs[id]; r != nil {
		_ = r.fsm.Drain(context.Background())
		_ = r.fsm.Stop(context.Background())
		r.cancel()
		delete(m.runs, id)
		m.reg.Unregister(id)
	}
}

// Reload применяет новый список доменов: стартует/перезапускает/останавливает.
func (m *DomainManager) Reload(ctx context.Context, specs []DomainSpec) {
	index := map[string]DomainSpec{}
	for _, s := range specs {
		if !m.manages(s) {
			if m.logger != nil {
				mode := s.Mode
				if mode == "" {
					mode = "inproc"
				}
				m.logger.Log(ctx, "DEBUG", "domain reload skipped", map[string]any{"id": s.ID, "mode": mode, "kind": s.Kind})
			}
			continue
		}
		index[s.ID] = s
	}

	// stop removed
	for id := range m.runs {
		if _, keep := index[id]; !keep {
			m.stop(id)
		}
	}
	// (re)start changed/new
	for id, s := range index {
		if r := m.runs[id]; r == nil {
			if err := m.launchInproc(ctx, s); err != nil && m.logger != nil {
				m.logger.Log(ctx, "ERROR", "domain reload launch failed", map[string]any{"id": s.ID, "kind": s.Kind, "err": err.Error()})
			}
			continue
		} else {
			// если изменились значимые поля — перезапуск
			old := r.spec
			oldFF := old.FeatureFlags
			newFF := s.FeatureFlags
			if old.Mode != s.Mode || old.Kind != s.Kind || !reflect.DeepEqual(old.Config, s.Config) || !reflect.DeepEqual(oldFF, newFF) {
				m.stop(id)
				if err := m.launchInproc(ctx, s); err != nil && m.logger != nil {
					m.logger.Log(ctx, "ERROR", "domain reload relaunch failed", map[string]any{"id": s.ID, "kind": s.Kind, "err": err.Error()})
				}
			} else {
				m.runs[id].spec = s
			}
		}
	}
}
