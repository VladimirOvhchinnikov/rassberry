module example.com/ffp/cmd/rk

go 1.22

require (
	example.com/ffp/platform/contracts v0.0.0
	example.com/ffp/platform/ports v0.0.0-00010101000000-000000000000
	example.com/ffp/platform/runtime v0.0.0
	example.com/ffp/platform/telemetry v0.0.0
	google.golang.org/grpc v1.66.0
	gopkg.in/yaml.v3 v3.0.1
)

replace example.com/ffp/platform/contracts => ../../platform/contracts

replace example.com/ffp/platform/runtime => ../../platform/runtime

replace example.com/ffp/platform/telemetry => ../../platform/telemetry

replace example.com/ffp/platform/ports => ../../platform/ports

replace google.golang.org/grpc => ../../third_party/google.golang.org/grpc

replace gopkg.in/yaml.v3 => ../../third_party/gopkg.in/yaml.v3
