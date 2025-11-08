package main

import (
	"sync"
	"time"

	"example.com/ffp/platform/contracts"
)

type KernelRecord struct {
	ID           string             `json:"id"`
	Scope        contracts.Scope    `json:"scope"`
	ParentID     string             `json:"parent_id"`
	Manifest     contracts.Manifest `json:"manifest"`
	Exports      *contracts.Exports `json:"exports,omitempty"`
	Health       contracts.Health   `json:"health"`
	RegisteredAt time.Time          `json:"registered_at"`
}

type DiscoveryRegistry struct {
	mu   sync.RWMutex
	data map[string]KernelRecord
}

func NewDiscoveryRegistry() *DiscoveryRegistry {
	return &DiscoveryRegistry{data: make(map[string]KernelRecord)}
}

func (r *DiscoveryRegistry) Register(rec KernelRecord) {
	r.mu.Lock()
	r.data[rec.ID] = rec
	r.mu.Unlock()
}

func (r *DiscoveryRegistry) Unregister(id string) {
	r.mu.Lock()
	delete(r.data, id)
	r.mu.Unlock()
}

func (r *DiscoveryRegistry) UpdateHealth(id string, h contracts.Health) {
	r.mu.Lock()
	rec, ok := r.data[id]
	if ok {
		rec.Health = h
		r.data[id] = rec
	}
	r.mu.Unlock()
}

func (r *DiscoveryRegistry) Get(id string) (KernelRecord, bool) {
	r.mu.RLock()
	rec, ok := r.data[id]
	r.mu.RUnlock()
	return rec, ok
}

func (r *DiscoveryRegistry) List() []KernelRecord {
	r.mu.RLock()
	out := make([]KernelRecord, 0, len(r.data))
	for _, v := range r.data {
		out = append(out, v)
	}
	r.mu.RUnlock()
	return out
}

// AggregateHealth возвращает сводный статус: failed > degraded > draining > ready > stopped.
func (r *DiscoveryRegistry) AggregateHealth() contracts.Health {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.data) == 0 {
		return contracts.Health{Status: contracts.HealthReady, Since: time.Now()}
	}

	status := contracts.HealthReady
	for _, rec := range r.data {
		switch rec.Health.Status {
		case contracts.HealthFailed:
			return contracts.Health{Status: contracts.HealthFailed, Since: rec.Health.Since, Reason: "some kernel failed"}
		case contracts.HealthDegraded:
			if status != contracts.HealthFailed {
				status = contracts.HealthDegraded
			}
		case contracts.HealthDraining:
			if status != contracts.HealthFailed && status != contracts.HealthDegraded {
				status = contracts.HealthDraining
			}
		}
	}
	return contracts.Health{Status: status, Since: time.Now()}
}
