package main

import "example.com/ffp/platform/contracts"

func (r *DiscoveryRegistry) SetExports(id string, ex *contracts.Exports) {
	r.mu.Lock()
	if rec, ok := r.kernels[id]; ok {
		rec.Exports = ex
	}
	r.mu.Unlock()
}
