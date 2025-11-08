module example.com/ffp/cmd/rkctl

go 1.22

require example.com/ffp/platform/telemetry v0.0.0

require google.golang.org/grpc v1.66.0 // indirect

replace example.com/ffp/platform/telemetry => ../../platform/telemetry

replace google.golang.org/grpc => ../../third_party/google.golang.org/grpc
