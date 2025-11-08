package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"example.com/ffp/platform/ports"
	"example.com/ffp/platform/telemetry"
)

func (s *AdminServer) AddLogStream(bus ports.EventBus) {
	mux, _ := s.srv.Handler.(*http.ServeMux)
	mux.HandleFunc("/admin/logs/stream", func(w http.ResponseWriter, r *http.Request) {
		// SSE заголовки
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		// парсим фильтры
		filter := telemetry.LogFilter{}
		if lv := r.URL.Query().Get("level"); lv != "" {
			switch lv {
			case "DEBUG", "debug":
				filter.LevelMin = telemetry.Debug
			case "WARN", "warn", "warning":
				filter.LevelMin = telemetry.Warn
			case "ERROR", "error":
				filter.LevelMin = telemetry.Error
			default:
				filter.LevelMin = telemetry.Info
			}
		}
		filter.KernelID = r.URL.Query().Get("kernel")
		filter.Scope = r.URL.Query().Get("scope")
		filter.Component = r.URL.Query().Get("component")

		ctx := r.Context()
		chAll, cancelAll, _ := bus.Subscribe(ctx, "telemetry.logs")
		defer cancelAll()

		// по желанию можно ещё и на scope-специфичную тему подписаться
		var chScope <-chan any
		var cancelScope func()
		if filter.Scope != "" {
			chScope, cancelScope, _ = bus.Subscribe(ctx, "telemetry.logs."+filter.Scope)
			defer func() {
				if cancelScope != nil {
					cancelScope()
				}
			}()
		}

		keep := time.NewTicker(10 * time.Second)
		defer keep.Stop()

		write := func(tag string, v any) bool {
			w.Write([]byte("event: " + tag + "\n"))
			b, _ := json.Marshal(v)
			w.Write([]byte("data: "))
			w.Write(b)
			w.Write([]byte("\n\n"))
			flusher.Flush()
			return true
		}

		_ = write("hello", map[string]string{"status": "ok", "ts": strconv.FormatInt(time.Now().Unix(), 10)})

		for {
			select {
			case <-ctx.Done():
				return
			case <-keep.C:
				w.Write([]byte(": keep-alive\n\n")) // комментарий SSE
				flusher.Flush()
			case m := <-chAll:
				if m == nil {
					continue
				}
				switch rec := m.(type) {
				case telemetry.LogRecordV2:
					if filter.MatchV2(rec) {
						_ = write("log", rec)
					}
				default:
					// игнорируем неизвестные типы
				}
			case m := <-chScope:
				if m == nil {
					continue
				}
				if rec, ok := m.(telemetry.LogRecordV2); ok && filter.MatchV2(rec) {
					_ = write("log", rec)
				}
			}
		}
	})
}
