package config

import "time"

// CacheConfig defines response caching behaviour for a route.
type CacheConfig struct {
	// Enabled turns caching on or off (default: false).
	Enabled bool `yaml:"enabled"`

	// TTL is how long a cached response is considered fresh.
	// Accepts Go duration strings, e.g. "30s", "5m".
	TTL string `yaml:"ttl"`

	// VaryByQuery lists query-parameter names that form part of the cache key.
	VaryByQuery []string `yaml:"vary_by_query"`

	// VaryByHeader lists request header names that form part of the cache key.
	VaryByHeader []string `yaml:"vary_by_header"`
}

// ParseTTL returns the parsed TTL duration or a sensible default (60s).
func (c *CacheConfig) ParseTTL() (time.Duration, error) {
	if c == nil || c.TTL == "" {
		return 60 * time.Second, nil
	}
	return time.ParseDuration(c.TTL)
}

// IsEnabled returns true when the cache config is non-nil and explicitly enabled.
func (c *CacheConfig) IsEnabled() bool {
	return c != nil && c.Enabled
}
