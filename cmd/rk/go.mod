module example.com/ffp/cmd/rk

go 1.22

require (
	example.com/ffp/platform/contracts v0.0.0
	example.com/ffp/platform/runtime v0.0.0
	example.com/ffp/platform/telemetry v0.0.0
)

replace example.com/ffp/platform/contracts => ../../platform/contracts

replace example.com/ffp/platform/runtime => ../../platform/runtime

replace example.com/ffp/platform/telemetry => ../../platform/telemetry
