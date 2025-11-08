package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"example.com/ffp/platform/ports"
)

type AdminServer struct {
	srv    *http.Server
	reg    *DiscoveryRegistry
	logger ports.Logger
	health *HealthAggregator
}

func NewAdminServer(addr string, reg *DiscoveryRegistry, logger ports.Logger) *AdminServer {
	mux := http.NewServeMux()
	srv := &http.Server{Addr: addr, Handler: mux}
	s := &AdminServer{srv: srv, reg: reg, logger: logger}
	s.registerBaseHandlers()
	return s
}

func (s *AdminServer) registerBaseHandlers() {
	mux, _ := s.srv.Handler.(*http.ServeMux)
	mux.HandleFunc("/admin/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		summary := s.reg.AggregateHealth()
		if s.health != nil {
			summary = s.health.Snapshot()
		}
		resp := map[string]any{
			"summary":      summary,
			"kernels":      s.reg.KernelHealth(),
			"generated_at": time.Now(),
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/admin/kernels", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(s.reg.Kernels())
	})
}

func (s *AdminServer) SetHealthAggregator(h *HealthAggregator) {
	s.health = h
}

func (s *AdminServer) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		err := s.srv.ListenAndServe()
		if err == http.ErrServerClosed {
			err = nil
		}
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.srv.Shutdown(shutdownCtx)
		<-errCh
		return nil
	case err := <-errCh:
		return err
	}
}
