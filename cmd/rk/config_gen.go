package main

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type AdminConfig struct {
	Addr string `yaml:"addr"`
}

type TelemetryConfig struct {
	Level  string `yaml:"level"`
	Buffer int    `yaml:"buffer"`
}

type DomainSpec struct {
	ID     string         `yaml:"id"`
	Mode   string         `yaml:"mode"`   // inproc|process|remote
	Kind   string         `yaml:"kind"`   // "example" или имя регистра
	Config map[string]any `yaml:"config"` // произвольные поля
}

type RootConfig struct {
	Admin     AdminConfig     `yaml:"admin"`
	Telemetry TelemetryConfig `yaml:"telemetry"`
	Domains   []DomainSpec    `yaml:"domains"`
}

func defaultConfig() RootConfig {
	return RootConfig{
		Admin:     AdminConfig{Addr: ":8090"},
		Telemetry: TelemetryConfig{Level: "INFO", Buffer: 256},
		Domains:   []DomainSpec{{ID: "example", Mode: "inproc", Kind: "example", Config: map[string]any{}}},
	}
}

func LoadConfig(path string) (RootConfig, error) {
	if path == "" {
		return defaultConfig(), nil
	}
	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return RootConfig{}, err
	}
	var cfg RootConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return RootConfig{}, err
	}
	if cfg.Admin.Addr == "" {
		return RootConfig{}, errors.New("admin.addr is required")
	}
	return cfg, nil
}
