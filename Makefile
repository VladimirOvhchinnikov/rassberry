SHELL := /bin/sh

.PHONY: build test tidy run.rk run.rkctl

build:
	@mkdir -p bin
	go build -o bin/rk ./cmd/rk
	go build -o bin/rkctl ./cmd/rkctl

test:
	go test ./...

tidy:
	go work sync
	( cd platform/contracts && go mod tidy )
	( cd platform/ports && go mod tidy )
	( cd platform/runtime && go mod tidy )
	( cd platform/telemetry && go mod tidy )
	( cd cmd/rk && go mod tidy )
	( cd cmd/rkctl && go mod tidy )

run.rk:
	go run ./cmd/rk

run.rkctl:
	go run ./cmd/rkctl
