package config

// CircuitBreakerConfig defines circuit breaker settings for a route.
// When the upstream error rate exceeds the threshold, the circuit opens
// and requests are rejected with the configured status code.
type CircuitBreakerConfig struct {
	// Enabled activates the circuit breaker for this route.
	Enabled bool `yaml:"enabled"`

	// Threshold is the number of consecutive failures before the circuit opens.
	// Defaults to 5 if not set.
	Threshold int `yaml:"threshold"`

	// OpenDuration is how long (in milliseconds) the circuit stays open
	// before moving to half-open. Defaults to 5000ms.
	OpenDuration int `yaml:"open_duration_ms"`

	// StatusCode is returned to the client when the circuit is open.
	// Defaults to 503.
	StatusCode int `yaml:"status_code"`

	// Body is the response body returned when the circuit is open.
	Body string `yaml:"body"`
}
