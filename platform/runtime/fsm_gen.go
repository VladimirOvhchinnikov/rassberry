package runtime

import (
	"context"
	"sync"
	"time"

	"example.com/ffp/platform/contracts"
)

// KernelModule — контракт ядра (минимальные hook-и). Можно встраивать KernelModuleBase.
type KernelModule interface {
	Manifest() contracts.Manifest
	OnLoad(ctx context.Context, host KernelHost) error
	OnInit(ctx context.Context) error
	OnConfigure(ctx context.Context, cfg map[string]any) error
	OnStart(ctx context.Context) error
	// OnReady вызывается после успешного Start; может быть no-op.
	OnReady(ctx context.Context) error
	OnDrain(ctx context.Context) error
	OnStop(ctx context.Context) error
	// Health возвращает текущий агрегированный статус ядра.
	Health() contracts.Health
}

// KernelModuleBase — no-op реализация hook-ов (для удобного встраивания).
type KernelModuleBase struct{}

func (KernelModuleBase) Manifest() contracts.Manifest                      { return contracts.Manifest{} }
func (KernelModuleBase) OnLoad(context.Context, KernelHost) error          { return nil }
func (KernelModuleBase) OnInit(context.Context) error                      { return nil }
func (KernelModuleBase) OnConfigure(context.Context, map[string]any) error { return nil }
func (KernelModuleBase) OnStart(context.Context) error                     { return nil }
func (KernelModuleBase) OnReady(context.Context) error                     { return nil }
func (KernelModuleBase) OnDrain(context.Context) error                     { return nil }
func (KernelModuleBase) OnStop(context.Context) error                      { return nil }
func (KernelModuleBase) Health() contracts.Health {
	return contracts.Health{Status: contracts.HealthStopped, Since: time.Now()}
}

// FSM — конечный автомат жизненного цикла ядра.
type FSM struct {
	mu     sync.RWMutex
	state  contracts.LifecycleState
	kernel KernelModule
	host   KernelHost
	on     func(from, to contracts.LifecycleState, err error)
}

type FSMOption func(*FSM)

// WithTransitionHook регистрирует callback на смену состояния.
func WithTransitionHook(h func(from, to contracts.LifecycleState, err error)) FSMOption {
	return func(f *FSM) { f.on = h }
}

func NewFSM(k KernelModule, h KernelHost, opts ...FSMOption) *FSM {
	f := &FSM{kernel: k, host: h, state: contracts.StateLoad}
	for _, o := range opts {
		o(f)
	}
	return f
}

func (f *FSM) State() contracts.LifecycleState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

func (f *FSM) set(to contracts.LifecycleState, err error) {
	f.mu.Lock()
	from := f.state
	f.state = to
	cb := f.on
	f.mu.Unlock()
	if cb != nil {
		cb(from, to, err)
	}
}

// Run проходит последовательность Load→Init→Configure→Start→Ready.
// При ошибке переводит в Failed и возвращает ошибку.
func (f *FSM) Run(ctx context.Context, cfg map[string]any) error {
	if err := f.kernel.OnLoad(ctx, f.host); err != nil {
		f.set(contracts.StateFailed, err)
		return err
	}
	f.set(contracts.StateInit, nil)

	if err := f.kernel.OnInit(ctx); err != nil {
		f.set(contracts.StateFailed, err)
		return err
	}
	f.set(contracts.StateConfigure, nil)

	if err := f.kernel.OnConfigure(ctx, cfg); err != nil {
		f.set(contracts.StateFailed, err)
		return err
	}
	f.set(contracts.StateStart, nil)

	if err := f.kernel.OnStart(ctx); err != nil {
		f.set(contracts.StateFailed, err)
		return err
	}

	// Успешный старт — ядро готово.
	f.set(contracts.StateReady, nil)

	// Доп. хук "OnReady" — необязательный этап.
	_ = f.kernel.OnReady(ctx)

	return nil
}

// Drain переводит ядро в Draining, вызывает OnDrain.
func (f *FSM) Drain(ctx context.Context) error {
	f.set(contracts.StateDraining, nil)
	if err := f.kernel.OnDrain(ctx); err != nil {
		f.set(contracts.StateFailed, err)
		return err
	}
	return nil
}

// Stop переводит в Stopped, вызывает OnStop.
func (f *FSM) Stop(ctx context.Context) error {
	if err := f.kernel.OnStop(ctx); err != nil {
		f.set(contracts.StateFailed, err)
		return err
	}
	f.set(contracts.StateStopped, nil)
	return nil
}
