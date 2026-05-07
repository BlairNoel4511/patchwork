package config

// RateLimit defines rate limiting configuration for a route.
type RateLimit struct {
	// RequestsPerSecond is the maximum number of requests allowed per second.
	RequestsPerSecond float64 `yaml:"requests_per_second"`
	// Burst is the maximum number of requests allowed to exceed the rate in a short period.
	Burst int `yaml:"burst"`
	// StatusCode is the HTTP status code returned when the rate limit is exceeded (default 429).
	StatusCode int `yaml:"status_code"`
	// Body is the response body returned when the rate limit is exceeded.
	Body string `yaml:"body"`
}
