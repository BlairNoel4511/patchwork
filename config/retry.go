package config

import "fmt"

// RetryConfig defines retry behaviour for a proxied or webhook route.
type RetryConfig struct {
	Attempts    int     `yaml:"attempts"`
	BackoffMS   int     `yaml:"backoff_ms"`
	Multiplier  float64 `yaml:"multiplier"`
	MaxBackoffMS int    `yaml:"max_backoff_ms"`
	RetryOn     []int   `yaml:"retry_on"` // HTTP status codes that trigger a retry
}

// Validate returns an error if the RetryConfig contains invalid values.
func (r *RetryConfig) Validate() error {
	if r == nil {
		return nil
	}
	if r.Attempts < 1 {
		return fmt.Errorf("retry attempts must be >= 1, got %d", r.Attempts)
	}
	if r.BackoffMS < 0 {
		return fmt.Errorf("retry backoff_ms must be >= 0, got %d", r.BackoffMS)
	}
	if r.Multiplier < 0 {
		return fmt.Errorf("retry multiplier must be >= 0, got %f", r.Multiplier)
	}
	if r.MaxBackoffMS < 0 {
		return fmt.Errorf("retry max_backoff_ms must be >= 0, got %d", r.MaxBackoffMS)
	}
	return nil
}

// IsEnabled reports whether retry is configured.
func (r *RetryConfig) IsEnabled() bool {
	return r != nil && r.Attempts > 1
}

// ShouldRetry reports whether the given HTTP status code warrants a retry.
// If RetryOn is empty, any 5xx status triggers a retry.
func (r *RetryConfig) ShouldRetry(status int) bool {
	if r == nil {
		return false
	}
	if len(r.RetryOn) == 0 {
		return status >= 500
	}
	for _, s := range r.RetryOn {
		if s == status {
			return true
		}
	}
	return false
}
