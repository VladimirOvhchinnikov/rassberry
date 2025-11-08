package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type RootSection struct {
	NodeID string `yaml:"node_id"`
	Zone   string `yaml:"zone"`
}

type DiscoveryConfig struct {
	Enabled           bool `yaml:"enabled"`
	AdvertiseInternal bool `yaml:"advertise_internal"`
}

type TelemetryFilters struct {
	Level     string `yaml:"level"`
	Kernel    string `yaml:"kernel"`
	Scope     string `yaml:"scope"`
	Component string `yaml:"component"`
}

type AdminConfig struct {
	Addr     string `yaml:"addr"`
	GRPCAddr string `yaml:"grpc_addr"`
}

type TelemetryConfig struct {
	Level   string           `yaml:"level"`
	Buffer  int              `yaml:"buffer"`
	Filters TelemetryFilters `yaml:"filters"`
}

type DomainSpec struct {
	ID           string          `yaml:"id"`
	Mode         string          `yaml:"mode"`
	Kind         string          `yaml:"kind"`
	Entry        string          `yaml:"entry"`   // для remote
	Command      string          `yaml:"command"` // для process
	FeatureFlags map[string]bool `yaml:"feature_flags"`
	Config       map[string]any  `yaml:"config"`
}

type RootConfig struct {
	Root      RootSection     `yaml:"root"`
	Admin     AdminConfig     `yaml:"admin"`
	Discovery DiscoveryConfig `yaml:"discovery"`
	Telemetry TelemetryConfig `yaml:"telemetry"`
	Domains   []DomainSpec    `yaml:"domains"`
}

func defaultConfig() RootConfig {
	return RootConfig{
		Root:      RootSection{NodeID: "rk-1", Zone: "dc-1"},
		Admin:     AdminConfig{Addr: ":8090", GRPCAddr: ":8079"},
		Discovery: DiscoveryConfig{Enabled: true, AdvertiseInternal: true},
		Telemetry: TelemetryConfig{Level: "INFO", Buffer: 256, Filters: TelemetryFilters{Level: "INFO"}},
		Domains:   []DomainSpec{{ID: "site", Mode: "inproc", Kind: "site", FeatureFlags: map[string]bool{"http": true, "workers": true, "log_forwarder": true}, Config: map[string]any{"http_addr": ":8081", "log_gateway": "127.0.0.1:8079"}}},
	}
}

func LoadConfig(path string) (RootConfig, error) {
	cfg := defaultConfig()
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (c RootConfig) Validate() error {
	if c.Admin.Addr == "" {
		return fmt.Errorf("admin.addr is required")
	}
	if c.Admin.GRPCAddr == "" {
		return fmt.Errorf("admin.grpc_addr is required")
	}
	return nil
}
