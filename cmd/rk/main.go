package main

import (
	"fmt"
	"time"

	"example.com/ffp/platform/contracts"
	rt "example.com/ffp/platform/runtime"
	"example.com/ffp/platform/telemetry"
)

func main() {
	manifest := contracts.Manifest{
		KernelID: "rk",
		Version:  "0.0.1",
		Scope:    contracts.RootScope,
		Features: []string{"admin", "discovery"},
	}

	log := telemetry.LogRecord{
		Time:     time.Now(),
		Level:    telemetry.Info,
		KernelID: "rk",
		Scope:    string(contracts.RootScope),
		Message:  "Root-Kernel: скелет запущен",
		Fields:   map[string]any{"go": rt.GoRuntimeVersion()},
	}

	fmt.Println(manifest.JSON())
	fmt.Println(log.String())
}
