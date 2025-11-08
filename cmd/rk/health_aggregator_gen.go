package main

import (
	"context"
	"sync/atomic"
	"time"

	"example.com/ffp/platform/contracts"
	"example.com/ffp/platform/ports"
)

type HealthAggregator struct {
	reg    *DiscoveryRegistry
	last   atomic.Value // contracts.Health
	bus    ports.EventBus
	logger ports.Logger
}

func NewHealthAggregator(reg *DiscoveryRegistry, bus ports.EventBus, logger ports.Logger) *HealthAggregator {
	h := &HealthAggregator{reg: reg, bus: bus, logger: logger}
	h.last.Store(contracts.Health{Status: contracts.HealthReady, Since: time.Now()})
	return h
}

func (h *HealthAggregator) Snapshot() contracts.Health {
	if v := h.last.Load(); v != nil {
		return v.(contracts.Health)
	}
	return contracts.Health{Status: contracts.HealthReady, Since: time.Now()}
}

func (h *HealthAggregator) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			v := h.reg.AggregateHealth()
			h.last.Store(v)
			// опционально можно публиковать в шину (пригодится позже)
			_ = h.bus.Publish(ctx, "telemetry.health.root", v)
		}
	}
}
