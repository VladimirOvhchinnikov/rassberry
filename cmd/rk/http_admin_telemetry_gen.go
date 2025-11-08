package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
)

var currentLogLevel atomic.Value // string: "DEBUG"/"INFO"/"WARN"/"ERROR"

func init() {
	currentLogLevel.Store("INFO")
}

func (s *AdminServer) AddTelemetryHandlers() {
	mux, _ := s.srv.Handler.(*http.ServeMux)
	mux.HandleFunc("/admin/telemetry", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Level string `json:"level"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		lv := strings.ToUpper(req.Level)
		switch lv {
		case "DEBUG", "INFO", "WARN", "ERROR":
		default:
			http.Error(w, "bad level", http.StatusBadRequest)
			return
		}
		currentLogLevel.Store(lv)
		_ = json.NewEncoder(w).Encode(map[string]string{"ok": "true", "level": lv})
	})
}
