package config

// RequestIDConfig controls automatic request ID injection.
type RequestIDConfig struct {
	// Enabled turns the feature on. Defaults to true if the block is present.
	Enabled *bool `yaml:"enabled"`

	// Header is the HTTP header name used to propagate the request ID.
	// Defaults to "X-Request-Id".
	Header string `yaml:"header"`

	// ForceNew always generates a new ID even when the incoming request
	// already carries one in the configured header.
	ForceNew bool `yaml:"force_new"`
}

// IsEnabled returns true when the config block is present and not explicitly
// disabled.
func (r *RequestIDConfig) IsEnabled() bool {
	if r == nil {
		return false
	}
	if r.Enabled != nil {
		return *r.Enabled
	}
	return true
}

// ResolvedHeader returns the header name, falling back to "X-Request-Id".
func (r *RequestIDConfig) ResolvedHeader() string {
	if r == nil || r.Header == "" {
		return "X-Request-Id"
	}
	return r.Header
}
