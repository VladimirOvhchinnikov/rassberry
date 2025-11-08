//go:build rk_run

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfgPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	configPath = *cfgPath

	cfg, err := LoadConfig(*cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := RunRootKernel(ctx, cfg); err != nil && err != context.Canceled {
		fmt.Fprintln(os.Stderr, "kernel error:", err)
		os.Exit(1)
	}
}
