package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"example.com/ffp/platform/contracts"
)

type AdminServer struct {
	addr string
	reg  *DiscoveryRegistry
	srv  *http.Server
}

func NewAdminServer(addr string, reg *DiscoveryRegistry) *AdminServer {
	mux := http.NewServeMux()

	s := &AdminServer{
		addr: addr,
		reg:  reg,
		srv:  &http.Server{Addr: addr, Handler: mux},
	}

	mux.HandleFunc("/admin/health", s.handleHealth)
	mux.HandleFunc("/admin/kernels", s.handleKernels)

	return s
}

func (s *AdminServer) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		shCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = s.srv.Shutdown(shCtx)
	}()
	return s.srv.ListenAndServe()
}

func (s *AdminServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := s.reg.AggregateHealth()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(health)
}

func (s *AdminServer) handleKernels(w http.ResponseWriter, r *http.Request) {
	list := s.reg.List()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

// helper: готовый "ready" health, когда нечего агрегировать
func readyHealth() contracts.Health {
	return contracts.Health{Status: contracts.HealthReady, Since: time.Now()}
}
