package main

import (
	"context"

	"example.com/ffp/platform/ports"
	rt "example.com/ffp/platform/runtime"
	"example.com/ffp/platform/telemetry"
)

type LogHub struct {
	bus ports.EventBus
}

func NewLogHub(bus ports.EventBus) *LogHub { return &LogHub{bus: bus} }

func (h *LogHub) Publish(ctx context.Context, rec telemetry.LogRecordV2) {
	if h == nil || h.bus == nil {
		return
	}
	_ = h.bus.Publish(ctx, rt.TopicTelemetryLogsAll, rec)
	_ = h.bus.Publish(ctx, rt.TopicTelemetryLogsRoot, rec)
	// при желании можно добавить scope-специфичные темы: telemetry.logs.<scope>
	if rec.Scope != "" {
		_ = h.bus.Publish(ctx, "telemetry.logs."+rec.Scope, rec)
	}
}
