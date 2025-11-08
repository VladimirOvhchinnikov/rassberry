//go:build rk_run

package main

import (
	"context"
	"flag"
	"fmt"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "путь к YAML конфигу Root-Kernel (по умолчанию — встроенный)")
	flag.Parse()

	if err := RunRootKernel(context.Background(), configPath); err != nil {
		fmt.Println("rk: error:", err)
	}
}
