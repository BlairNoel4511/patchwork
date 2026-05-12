package config

import "fmt"

// APIKeyConfig defines API key authentication for a route.
type APIKeyConfig struct {
	// Keys is the list of accepted API key values.
	Keys []string `yaml:"keys"`

	// Header is the HTTP header name to read the key from (e.g. "X-API-Key").
	Header string `yaml:"header"`

	// QueryParam is the URL query parameter name to read the key from.
	QueryParam string `yaml:"query_param"`

	// Status is the HTTP status code returned on failure (default 401).
	Status *int `yaml:"status"`

	// Body is the response body returned on failure.
	Body string `yaml:"body"`
}

// IsEnabled returns true when there is at least one key and at least one
// source (header or query param) configured.
func (a *APIKeyConfig) IsEnabled() bool {
	if a == nil {
		return false
	}
	return len(a.Keys) > 0 && (a.Header != "" || a.QueryParam != "")
}

// ResolvedStatus returns the configured failure status or 401.
func (a *APIKeyConfig) ResolvedStatus() int {
	if a.Status != nil {
		return *a.Status
	}
	return 401
}

// ResolvedBody returns the configured failure body or a default JSON message.
func (a *APIKeyConfig) ResolvedBody(reason string) string {
	if a.Body != "" {
		return a.Body
	}
	return fmt.Sprintf(`{"error":%q}`, reason)
}
