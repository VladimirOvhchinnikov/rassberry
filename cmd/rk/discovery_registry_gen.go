package main

import (
	"sort"
	"sync"
	"time"

	"example.com/ffp/platform/contracts"
)

type KernelRecord struct {
	Manifest  contracts.Manifest `json:"manifest"`
	Health    contracts.Health   `json:"health"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type DiscoveryRegistry struct {
	mu      sync.RWMutex
	kernels map[string]*KernelRecord
}

func NewDiscoveryRegistry() *DiscoveryRegistry {
	return &DiscoveryRegistry{kernels: make(map[string]*KernelRecord)}
}

func (r *DiscoveryRegistry) RegisterKernel(m contracts.Manifest) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec, ok := r.kernels[m.KernelID]
	if !ok {
		rec = &KernelRecord{}
		r.kernels[m.KernelID] = rec
	}
	rec.Manifest = m
	if rec.Health.Status == "" {
		rec.Health = contracts.Health{Status: contracts.HealthReady, Since: time.Now()}
	}
	rec.UpdatedAt = time.Now()
}

func (r *DiscoveryRegistry) UpdateHealth(id string, health contracts.Health) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec, ok := r.kernels[id]
	if !ok {
		rec = &KernelRecord{Manifest: contracts.Manifest{KernelID: id}}
		r.kernels[id] = rec
	}
	rec.Health = health
	if rec.Health.Since.IsZero() {
		rec.Health.Since = time.Now()
	}
	rec.UpdatedAt = time.Now()
}

func (r *DiscoveryRegistry) Kernels() []KernelRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]KernelRecord, 0, len(r.kernels))
	for _, rec := range r.kernels {
		out = append(out, *rec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Manifest.KernelID < out[j].Manifest.KernelID })
	return out
}

func (r *DiscoveryRegistry) KernelHealth() map[string]contracts.Health {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make(map[string]contracts.Health, len(r.kernels))
	for id, rec := range r.kernels {
		res[id] = rec.Health
	}
	return res
}

func (r *DiscoveryRegistry) AggregateHealth() contracts.Health {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.kernels) == 0 {
		return contracts.Health{Status: contracts.HealthReady, Since: time.Now()}
	}
	summary := contracts.Health{Status: contracts.HealthReady, Since: time.Now()}
	var reason string
	for _, rec := range r.kernels {
		h := rec.Health
		if h.Status == "" {
			continue
		}
		if summary.Since.IsZero() || h.Since.Before(summary.Since) {
			summary.Since = h.Since
		}
		switch h.Status {
		case contracts.HealthFailed:
			summary.Status = contracts.HealthFailed
			if reason == "" {
				reason = "kernel failed: " + rec.Manifest.KernelID
			}
		case contracts.HealthDegraded:
			if summary.Status != contracts.HealthFailed {
				summary.Status = contracts.HealthDegraded
				if reason == "" {
					reason = "kernel degraded: " + rec.Manifest.KernelID
				}
			}
		case contracts.HealthDraining:
			if summary.Status != contracts.HealthFailed && summary.Status != contracts.HealthDegraded {
				summary.Status = contracts.HealthDraining
				if reason == "" {
					reason = "kernel draining: " + rec.Manifest.KernelID
				}
			}
		case contracts.HealthStopped:
			if summary.Status == contracts.HealthReady {
				summary.Status = contracts.HealthStopped
				if reason == "" {
					reason = "kernel stopped: " + rec.Manifest.KernelID
				}
			}
		default:
			if summary.Status == "" {
				summary.Status = contracts.HealthReady
			}
		}
	}
	if summary.Since.IsZero() {
		summary.Since = time.Now()
	}
	summary.Reason = reason
	if summary.Status == "" {
		summary.Status = contracts.HealthReady
	}
	return summary
}
