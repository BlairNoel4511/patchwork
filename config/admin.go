package config

// AdminConfig controls the built-in admin API endpoints.
type AdminConfig struct {
	// Enabled toggles the admin API (default: true).
	Enabled *bool `yaml:"enabled"`
	// Prefix is the URL path prefix for admin routes (default: "/__admin").
	Prefix string `yaml:"prefix"`
	// ReadOnly disables mutating operations (scenario set, log reset, etc.).
	ReadOnly bool `yaml:"read_only"`
}

// AdminPrefix returns the effective prefix, falling back to "/__admin".
func (a *AdminConfig) AdminPrefix() string {
	if a == nil || a.Prefix == "" {
		return "/__admin"
	}
	return a.Prefix
}

// IsEnabled returns true unless explicitly disabled.
func (a *AdminConfig) IsEnabled() bool {
	if a == nil || a.Enabled == nil {
		return true
	}
	return *a.Enabled
}
