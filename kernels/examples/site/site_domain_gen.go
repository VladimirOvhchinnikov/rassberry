package site

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"example.com/ffp/kernels/infra/log-forwarder"
	"example.com/ffp/platform/contracts"
	"example.com/ffp/platform/ports"
	rt "example.com/ffp/platform/runtime"
)

type Domain struct {
	rt.KernelModuleBase

	id       string
	host     rt.KernelHost
	logger   ports.Logger
	sup      *rt.Supervisor
	httpAddr string
	logGW    string

	health contracts.Health
}

// NewDomain создаёт доменное ядро "site".
func NewDomain(id string) *Domain {
	return &Domain{
		id:     id,
		health: contracts.Health{Status: contracts.HealthReady, Since: time.Now()},
		// значения по умолчанию
		httpAddr: ":8081",
		logGW:    "127.0.0.1:8079",
	}
}

func (d *Domain) Manifest() contracts.Manifest {
	return contracts.Manifest{
		KernelID: d.id, Version: "0.0.1", Scope: contracts.DomainScope,
		Features: []string{"http", "workers", "log-forwarder"},
	}
}

func (d *Domain) OnLoad(ctx context.Context, host rt.KernelHost) error {
	d.host = host
	d.logger = host.Logger()
	return nil
}

func (d *Domain) OnConfigure(ctx context.Context, cfg map[string]any) error {
	if v, ok := cfg["http_addr"].(string); ok && v != "" {
		d.httpAddr = v
	}
	if v, ok := cfg["log_gateway"].(string); ok && v != "" {
		d.logGW = v
	}
	return nil
}

func (d *Domain) OnStart(ctx context.Context) error {
	d.sup = rt.NewSupervisor(rt.WithOnEvent(func(ev rt.Event) {
		// при желании можно логировать события супервизора
	}))

	// FK: hello (HTTP GET /hello)
	d.sup.Start(rt.WorkerSpec{
		Name:   "fk-hello-http",
		Policy: rt.Permanent,
		Fn: func(ctx context.Context) error {
			mux := http.NewServeMux()
			mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "hello from site: %s\n", d.id)
			})
			srv := &http.Server{Addr: d.httpAddr, Handler: mux}
			go func() {
				<-ctx.Done()
				_ = srv.Shutdown(context.Background())
			}()
			d.logger.Log(ctx, "INFO", "http hello listening", map[string]any{"addr": d.httpAddr})
			err := srv.ListenAndServe()
			if err == http.ErrServerClosed {
				return nil
			}
			return err
		},
	})

	// FK: log-forwarder — пересылка логов домена в Root LogGateway
	d.sup.Start(rt.WorkerSpec{
		Name:   "fk-log-forwarder",
		Policy: rt.Permanent,
		Fn: func(ctx context.Context) error {
			fwd := logforwarder.New(d.logGW, d.host.EventBus(), d.logger)
			return fwd.Run(ctx) // блокирует до ctx.Done()
		},
	})

	// Воркер: периодическая генерация тест-логов
	d.sup.Start(rt.WorkerSpec{
		Name:   "worker-logs",
		Policy: rt.Permanent,
		Fn: func(ctx context.Context) error {
			t := time.NewTicker(3 * time.Second)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return nil
				case <-t.C:
					d.logger.Log(ctx, "INFO", "site heartbeat", map[string]any{"component": "site/heartbeat"})
				}
			}
		},
	})

	return nil
}

func (d *Domain) OnStop(ctx context.Context) error {
	if d.sup != nil {
		d.sup.StopAll()
		d.sup.Wait()
	}
	return nil
}

func (d *Domain) Health() contracts.Health { return d.health }
