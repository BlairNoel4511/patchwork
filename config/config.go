package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Route defines a single mock HTTP route.
type Route struct {
	Path        string            `yaml:"path"`
	Method      string            `yaml:"method"`
	Status      int               `yaml:"status"`
	Body        string            `yaml:"body"`
	Headers     map[string]string `yaml:"headers"`
	DelayMs     int               `yaml:"delay_ms"`
}

// Config holds the full server configuration.
type Config struct {
	Port   int     `yaml:"port"`
	Routes []Route `yaml:"routes"`
}

// Load reads and parses a YAML config file from the given path.
// It applies sensible defaults where values are omitted.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Port == 0 {
		cfg.Port = 8080
	}

	for i := range cfg.Routes {
		if cfg.Routes[i].Method == "" {
			cfg.Routes[i].Method = "GET"
		}
		if cfg.Routes[i].Status == 0 {
			cfg.Routes[i].Status = 200
		}
	}

	return &cfg, nil
}
