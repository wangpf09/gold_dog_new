package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var cfg *Config

func GetConfig() *Config {
	return cfg
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.QOSConfig.validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	return nil
}
