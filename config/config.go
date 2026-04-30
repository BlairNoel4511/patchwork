package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Route defines a single mock HTTP route.
type Route struct {
	Method   string            `yaml:"method"`
	Path     string            `yaml:"path"`
	Status   int               `yaml:"status"`
	Headers  map[string]string `yaml:"headers"`
	Body     string            `yaml:"body"`
}

// Config is the top-level structure of the YAML config file.
type Config struct {
	Port   int     `yaml:"port"`
	Routes []Route `yaml:"routes"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Port == 0 {
		cfg.Port = 8080
	}

	for i, r := range cfg.Routes {
		if r.Status == 0 {
			cfg.Routes[i].Status = 200
		}
		if r.Method == "" {
			cfg.Routes[i].Method = "GET"
		}
	}

	return &cfg, nil
}
