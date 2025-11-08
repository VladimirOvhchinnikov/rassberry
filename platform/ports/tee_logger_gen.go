package ports

import (
	"context"
	"strings"
	"time"

	"example.com/ffp/platform/telemetry"
)

type teeLogger struct {
	bus       EventBus
	kernelID  string
	scope     string
	component string
	next      Logger
}

// NewTeeLogger создаёт логгер, публикующий записи в EventBus и (опционально) делегирующий следующему логгеру.
func NewTeeLogger(bus EventBus, kernelID, scope, component string, delegates ...Logger) Logger {
	var next Logger
	if len(delegates) > 0 {
		next = delegates[0]
	}
	return &teeLogger{bus: bus, kernelID: kernelID, scope: scope, component: component, next: next}
}

func (l *teeLogger) Log(ctx context.Context, level string, message string, fields map[string]any) {
	if l == nil {
		return
	}
	if l.next != nil {
		l.next.Log(ctx, level, message, fields)
	}
	lvl := telemetry.Info
	switch strings.ToUpper(level) {
	case "DEBUG":
		lvl = telemetry.Debug
	case "INFO":
		lvl = telemetry.Info
	case "WARN", "WARNING":
		lvl = telemetry.Warn
	case "ERROR":
		lvl = telemetry.Error
	}
	rec := telemetry.LogRecordV2{
		Time:      time.Now(),
		Level:     lvl,
		KernelID:  l.kernelID,
		Scope:     l.scope,
		Component: l.component,
		Message:   message,
	}
	if len(fields) > 0 {
		copy := make(map[string]any, len(fields))
		for k, v := range fields {
			copy[k] = v
		}
		rec.Fields = copy
	}
	if l.bus == nil {
		return
	}
	topics := []string{"telemetry.logs"}
	switch strings.ToLower(l.scope) {
	case "root":
		topics = append(topics, "telemetry.logs.root")
	case "domain":
		topics = append(topics, "telemetry.logs.domain")
	case "function":
		topics = append(topics, "telemetry.logs.function")
	}
	for _, topic := range topics {
		_ = l.bus.Publish(ctx, topic, rec)
	}
}
