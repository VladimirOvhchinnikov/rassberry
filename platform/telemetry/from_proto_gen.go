package telemetry

import (
	"strings"
	"time"

	telemetrypb "example.com/ffp/platform/telemetry/proto"
)

// FromProto преобразует protobuf-запрос в LogRecordV2.
func FromProto(req *telemetrypb.PushRequest) LogRecordV2 {
	rec := LogRecordV2{}
	if req == nil {
		rec.Time = time.Now()
		rec.Level = Info
		return rec
	}
	r := req.Record
	if r == nil {
		rec.Time = time.Now()
		rec.Level = Info
		return rec
	}
	if r.TimeUnixNano > 0 {
		rec.Time = time.Unix(0, r.TimeUnixNano)
	} else {
		rec.Time = time.Now()
	}
	rec.Level = parseLevel(r.Level)
	rec.KernelID = r.KernelId
	rec.Scope = r.Scope
	rec.Component = r.Component
	rec.Trace = r.Trace
	rec.Message = r.Message
	if len(r.Fields) > 0 {
		rec.Fields = make(map[string]any, len(r.Fields))
		for _, f := range r.Fields {
			if f == nil {
				continue
			}
			rec.Fields[f.Key] = f.Value
		}
	}
	return rec
}

func parseLevel(lv string) Level {
	switch strings.ToUpper(lv) {
	case "DEBUG":
		return Debug
	case "WARN", "WARNING":
		return Warn
	case "ERROR":
		return Error
	default:
		return Info
	}
}
