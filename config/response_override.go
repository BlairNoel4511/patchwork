package config

// ResponseOverride defines a runtime override for a named route's response.
// When active, all requests to the matching route will receive the overridden
// status code and/or body until the override is cleared.
type ResponseOverride struct {
	// Status overrides the HTTP status code. Zero means no override.
	Status int `yaml:"status,omitempty"`
	// Body overrides the response body. Empty string means no override.
	Body string `yaml:"body,omitempty"`
	// Headers are additional headers to inject when the override is active.
	Headers map[string]string `yaml:"headers,omitempty"`
}

// IsEmpty reports whether the override carries no meaningful values.
func (r *ResponseOverride) IsEmpty() bool {
	if r == nil {
		return true
	}
	return r.Status == 0 && r.Body == "" && len(r.Headers) == 0
}
