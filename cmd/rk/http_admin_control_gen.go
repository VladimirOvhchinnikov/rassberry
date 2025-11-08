package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"example.com/ffp/platform/contracts"
)

func (s *AdminServer) AddKernelControlHandlers() {
	mux, _ := s.srv.Handler.(*http.ServeMux)
	mux.HandleFunc("/admin/kernels/", func(w http.ResponseWriter, r *http.Request) {
		// /admin/kernels/{id}/restart | /drain
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/kernels/"), "/")
		if len(parts) < 2 {
			http.Error(w, "bad path", http.StatusBadRequest)
			return
		}
		id, action := parts[0], parts[1]
		switch r.Method {
		case http.MethodPost:
			switch action {
			case "restart":
				// Демоверсия: помечаем health как ready
				s.reg.UpdateHealth(id, contracts.Health{Status: contracts.HealthReady, Since: time.Now(), Reason: "manual restart (placeholder)"})
				w.WriteHeader(http.StatusAccepted)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "accepted", "action": "restart", "id": id})
			case "drain":
				s.reg.UpdateHealth(id, contracts.Health{Status: contracts.HealthDraining, Since: time.Now(), Reason: "manual drain (placeholder)"})
				w.WriteHeader(http.StatusAccepted)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "accepted", "action": "drain", "id": id})
			default:
				http.Error(w, "unknown action", http.StatusNotFound)
			}
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
