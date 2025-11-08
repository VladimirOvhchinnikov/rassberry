package ports

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type teeLogger struct {
	mu        sync.Mutex
	bus       EventBus
	kernelID  string
	scope     string
	component string
}

// NewTeeLogger создаёт логгер, который пишет в stdout и публикует записи в EventBus (если доступен).
func NewTeeLogger(bus EventBus, kernelID, scope, component string) Logger {
	return &teeLogger{bus: bus, kernelID: kernelID, scope: scope, component: component}
}

func (l *teeLogger) Log(ctx context.Context, level string, message string, fields map[string]any) {
	if fields == nil {
		fields = map[string]any{}
	}
	record := map[string]any{
		"time":      time.Now().UTC(),
		"level":     level,
		"kernel_id": l.kernelID,
		"scope":     l.scope,
		"component": l.component,
		"message":   message,
		"fields":    fields,
	}

	l.mu.Lock()
	fmt.Printf("%s %-5s %s/%s %s %v\n", time.Now().Format(time.RFC3339), level, l.scope, l.kernelID, message, fields)
	l.mu.Unlock()

	if l.bus != nil {
		_ = l.bus.Publish(ctx, "telemetry.logs", record)
	}
}
