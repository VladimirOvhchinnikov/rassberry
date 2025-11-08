package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type RootConfig struct {
	Admin AdminConfig `yaml:"admin"`
}

type AdminConfig struct {
	Addr     string `yaml:"addr"`
	GRPCAddr string `yaml:"grpc_addr"`
}

func defaultConfig() RootConfig {
	return RootConfig{
		Admin: AdminConfig{Addr: ":8090", GRPCAddr: ":8079"},
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
