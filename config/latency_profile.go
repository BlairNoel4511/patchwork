package config

// LatencyProfile defines a named latency distribution for a route.
type LatencyProfile struct {
	// Distribution is one of: fixed, uniform, normal
	Distribution string  `yaml:"distribution"`
	FixedMs      int     `yaml:"fixed_ms"`
	MinMs        int     `yaml:"min_ms"`
	MaxMs        int     `yaml:"max_ms"`
	MeanMs       float64 `yaml:"mean_ms"`
	StdDevMs     float64 `yaml:"std_dev_ms"`
}
