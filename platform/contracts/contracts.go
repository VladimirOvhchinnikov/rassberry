package contracts

import (
	"encoding/json"
	"time"
)

// Scope описывает уровень ядра.
type Scope string

const (
	RootScope     Scope = "root"
	DomainScope   Scope = "domain"
	FunctionScope Scope = "function"
)

// LifecycleState фиксирует этап жизненного цикла ядра.
type LifecycleState string

const (
	StateLoad      LifecycleState = "load"
	StateInit      LifecycleState = "init"
	StateConfigure LifecycleState = "configure"
	StateStart     LifecycleState = "start"
	StateReady     LifecycleState = "ready"
	StateDegraded  LifecycleState = "degraded"
	StateFailed    LifecycleState = "failed"
	StateDraining  LifecycleState = "draining"
	StateStopped   LifecycleState = "stopped"
)

// HealthStatus — агрегированный статус здоровья.
type HealthStatus string

const (
	HealthReady    HealthStatus = "ready"
	HealthDegraded HealthStatus = "degraded"
	HealthFailed   HealthStatus = "failed"
	HealthDraining HealthStatus = "draining"
	HealthStopped  HealthStatus = "stopped"
)

// Health — минимальная модель health-отчёта.
type Health struct {
	Status HealthStatus `json:"status"`
	Reason string       `json:"reason,omitempty"`
	Since  time.Time    `json:"since"`
}

// Manifest — паспорт ядра.
type Manifest struct {
	KernelID  string         `json:"kernel_id"`
	Version   string         `json:"version"`
	Scope     Scope          `json:"scope"`
	Features  []string       `json:"features,omitempty"`
	Requires  map[string]any `json:"requires,omitempty"`
	Resources map[string]any `json:"resources,omitempty"`
	Security  map[string]any `json:"security,omitempty"`
	Compat    map[string]any `json:"compat,omitempty"`
}

// JSON возвращает человекочитаемое представление манифеста.
func (m Manifest) JSON() string {
	b, _ := json.MarshalIndent(m, "", "  ")
	return string(b)
}

// Envelope — общий конверт сообщения для портов/транспорта.
type Envelope struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Key        string            `json:"key,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Payload    []byte            `json:"payload,omitempty"`
	OccurredAt time.Time         `json:"occurred_at"`
}
