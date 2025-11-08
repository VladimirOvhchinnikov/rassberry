package runtime

import (
	"context"

	"example.com/ffp/platform/contracts"
	"example.com/ffp/platform/ports"
)

// KernelHost — минимальный DI-контейнер, предоставляемый ядру.
type KernelHost interface {
	ID() string
	Scope() contracts.Scope
	Logger() ports.Logger
	RPC() ports.RPC
	EventBus() ports.EventBus
	Stream() ports.Stream
	Config() map[string]any
}

type host struct {
	id     string
	scope  contracts.Scope
	logger ports.Logger
	rpc    ports.RPC
	bus    ports.EventBus
	stream ports.Stream
	cfg    map[string]any
}

// NewHost создаёт KernelHost. Все поля опциональны, но Logger по умолчанию — noop.
func NewHost(id string, scope contracts.Scope, opts ...HostOption) KernelHost {
	h := &host{
		id:     id,
		scope:  scope,
		cfg:    map[string]any{},
		logger: noopLogger{}, // безопасный дефолт
	}
	for _, o := range opts {
		o(h)
	}
	return h
}

type HostOption func(*host)

func WithLogger(l ports.Logger) HostOption {
	return func(h *host) {
		if l != nil {
			h.logger = l
		}
	}
}
func WithRPC(r ports.RPC) HostOption           { return func(h *host) { h.rpc = r } }
func WithEventBus(b ports.EventBus) HostOption { return func(h *host) { h.bus = b } }
func WithStream(s ports.Stream) HostOption     { return func(h *host) { h.stream = s } }
func WithConfig(cfg map[string]any) HostOption {
	return func(h *host) {
		if cfg == nil {
			return
		}
		for k, v := range cfg {
			h.cfg[k] = v
		}
	}
}

func (h *host) ID() string               { return h.id }
func (h *host) Scope() contracts.Scope   { return h.scope }
func (h *host) Logger() ports.Logger     { return h.logger }
func (h *host) RPC() ports.RPC           { return h.rpc }
func (h *host) EventBus() ports.EventBus { return h.bus }
func (h *host) Stream() ports.Stream     { return h.stream }
func (h *host) Config() map[string]any   { return h.cfg }

// noopLogger — безопасная заглушка.
type noopLogger struct{}

func (noopLogger) Log(context.Context, string, string, map[string]any) {}
