package config

import "strings"

// CORSConfig defines cross-origin resource sharing settings for a route or globally.
type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"` // seconds
}

// IsEnabled returns true when a CORSConfig is present and has at least one allowed origin.
func (c *CORSConfig) IsEnabled() bool {
	return c != nil && len(c.AllowedOrigins) > 0
}

// AllowedOriginsValue returns the joined allowed origins or "*" when the list
// contains a wildcard entry.
func (c *CORSConfig) AllowedOriginsValue() string {
	if c == nil {
		return ""
	}
	for _, o := range c.AllowedOrigins {
		if o == "*" {
			return "*"
		}
	}
	return strings.Join(c.AllowedOrigins, ", ")
}

// AllowedMethodsValue returns the methods as a comma-separated string,
// defaulting to common safe methods when none are configured.
func (c *CORSConfig) AllowedMethodsValue() string {
	if c == nil || len(c.AllowedMethods) == 0 {
		return "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	}
	return strings.Join(c.AllowedMethods, ", ")
}

// AllowedHeadersValue returns the headers as a comma-separated string,
// defaulting to Content-Type and Authorization when none are configured.
func (c *CORSConfig) AllowedHeadersValue() string {
	if c == nil || len(c.AllowedHeaders) == 0 {
		return "Content-Type, Authorization"
	}
	return strings.Join(c.AllowedHeaders, ", ")
}

// ExposedHeadersValue returns the exposed headers as a comma-separated string,
// or an empty string when none are configured.
func (c *CORSConfig) ExposedHeadersValue() string {
	if c == nil || len(c.ExposedHeaders) == 0 {
		return ""
	}
	return strings.Join(c.ExposedHeaders, ", ")
}
