SHELL := /bin/sh

.PHONY: build test tidy run.rk run.rkctl lint e2e

build:
	@mkdir -p bin
	go build -o bin/rk ./cmd/rk
	go build -o bin/rkctl ./cmd/rkctl

test:
	go test ./...

lint:
	go vet ./...; \
	gofmt -s -l . | (! grep .) || (echo "gofmt needed"; exit 1)

tidy:
	go work sync
	( cd platform/contracts && go mod tidy )
	( cd platform/ports && go mod tidy )
	( cd platform/runtime && go mod tidy )
	( cd platform/telemetry && go mod tidy )
	( cd cmd/rk && go mod tidy )
	( cd cmd/rkctl && go mod tidy )

run.rk:
	go run -tags rk_run ./cmd/rk -config ./cmd/rk/config.sample.yaml

run.rkctl:
	go run -tags rkctl_run ./cmd/rkctl

e2e:
	go test -tags e2e ./tests/e2e -v
