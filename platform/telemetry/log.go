package telemetry

import (
	"fmt"
	"time"
)

// Level — уровень логирования.
type Level int

const (
	Debug Level = iota
	Info
	Warn
	Error
)

func (l Level) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LogRecord — унифицированная запись лога.
type LogRecord struct {
	Time      time.Time      `json:"time"`
	Level     Level          `json:"level"`
	KernelID  string         `json:"kernel_id"`
	Scope     string         `json:"scope"`
	Component string         `json:"component,omitempty"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

func (r LogRecord) String() string {
	return fmt.Sprintf("%s %-5s %s/%s %s %v",
		r.Time.Format(time.RFC3339),
		r.Level.String(),
		r.Scope, r.KernelID,
		r.Message,
		r.Fields,
	)
}
