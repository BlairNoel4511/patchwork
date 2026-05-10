package config

import "errors"

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

// Validate checks that the CircuitBreakerConfig has valid field values.
// It returns an error if any field is out of an acceptable range.
func (c *CircuitBreakerConfig) Validate() error {
	if c.Threshold < 0 {
		return errors.New("circuit breaker threshold must be non-negative")
	}
	if c.OpenDuration < 0 {
		return errors.New("circuit breaker open_duration_ms must be non-negative")
	}
	if c.StatusCode != 0 && (c.StatusCode < 100 || c.StatusCode > 599) {
		return errors.New("circuit breaker status_code must be a valid HTTP status code (100-599)")
	}
	return nil
}
