package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port   int     `yaml:"port"`
	Routes []Route `yaml:"routes"`
}

type Route struct {
	Path      string     `yaml:"path"`
	Method    string     `yaml:"method"`
	Status    int        `yaml:"status"`
	Body      string     `yaml:"body"`
	Headers   map[string]string `yaml:"headers"`
	Delay     string     `yaml:"delay"`
	Responses []Response `yaml:"responses"`
	Conditions []Condition `yaml:"conditions"`
	Chaos     *ChaosConfig `yaml:"chaos"`
}

type Response struct {
	Status  int               `yaml:"status"`
	Body    string            `yaml:"body"`
	Headers map[string]string `yaml:"headers"`
}

type Condition struct {
	Match    MatchCondition `yaml:"match"`
	Status   int           `yaml:"status"`
	Body     string        `yaml:"body"`
	Headers  map[string]string `yaml:"headers"`
}

type MatchCondition struct {
	Header map[string]string `yaml:"header"`
	Query  map[string]string `yaml:"query"`
	Body   map[string]string `yaml:"body"`
}

type ChaosConfig struct {
	ErrorRate  float64 `yaml:"error_rate"`
	StatusCode int     `yaml:"status_code"`
	Body       string  `yaml:"body"`
}

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
