package main

import (
	"context"
	"sync"
	"time"

	"example.com/ffp/platform/contracts"
)

type DegradationPolicy struct {
	reg      *DiscoveryRegistry
	mu       sync.Mutex
	snapshot map[string]*contracts.Exports // оригинальные экспорты
}

func NewDegradationPolicy(reg *DiscoveryRegistry) *DegradationPolicy {
	return &DegradationPolicy{reg: reg, snapshot: make(map[string]*contracts.Exports)}
}

func (p *DegradationPolicy) copyExports(ex *contracts.Exports) *contracts.Exports {
	if ex == nil {
		return nil
	}
	cp := *ex
	if ex.Network != nil {
		cp.Network = append([]contracts.NetworkEndpoint(nil), ex.Network...)
	}
	if ex.Events != nil {
		cp.Events = append([]contracts.EventSpec(nil), ex.Events...)
	}
	if ex.Streams != nil {
		cp.Streams = append([]contracts.StreamSpec(nil), ex.Streams...)
	}
	if ex.CLI != nil {
		cp.CLI = append([]contracts.CLICommand(nil), ex.CLI...)
	}
	if ex.Local != nil {
		cp.Local = append([]contracts.LocalService(nil), ex.Local...)
	}
	return &cp
}

func (p *DegradationPolicy) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Second
	}
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			list := p.reg.List()
			for _, rec := range list {
				switch rec.Health.Status {
				case contracts.HealthReady:
					// восстановить экспорты, если были скрыты
					p.mu.Lock()
					if ex, ok := p.snapshot[rec.ID]; ok {
						p.reg.SetExports(rec.ID, ex)
						delete(p.snapshot, rec.ID)
					}
					p.mu.Unlock()
				case contracts.HealthDegraded, contracts.HealthFailed, contracts.HealthDraining, contracts.HealthStopped:
					// сохранить и скрыть экспорты
					if rec.Exports != nil {
						p.mu.Lock()
						if _, ok := p.snapshot[rec.ID]; !ok {
							p.snapshot[rec.ID] = p.copyExports(rec.Exports)
						}
						p.mu.Unlock()
						p.reg.SetExports(rec.ID, nil)
					}
				}
			}
		}
	}
}
