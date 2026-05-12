package config

// TransformConfig defines request/response transformation rules for a route.
type TransformConfig struct {
	// Request transformations applied before proxying or handling
	Request *RequestTransform `yaml:"request,omitempty"`
	// Response transformations applied before returning to client
	Response *ResponseTransform `yaml:"response,omitempty"`
}

// RequestTransform defines mutations to apply to incoming requests.
type RequestTransform struct {
	// StripPathPrefix removes a prefix from the request path
	StripPathPrefix string `yaml:"strip_path_prefix,omitempty"`
	// AddPathPrefix prepends a prefix to the request path
	AddPathPrefix string `yaml:"add_path_prefix,omitempty"`
	// SetHeaders forcibly sets headers on the forwarded request
	SetHeaders map[string]string `yaml:"set_headers,omitempty"`
	// RemoveHeaders strips headers from the forwarded request
	RemoveHeaders []string `yaml:"remove_headers,omitempty"`
}

// ResponseTransform defines mutations to apply to outgoing responses.
type ResponseTransform struct {
	// SetHeaders forcibly sets headers on the response
	SetHeaders map[string]string `yaml:"set_headers,omitempty"`
	// RemoveHeaders strips headers from the response
	RemoveHeaders []string `yaml:"remove_headers,omitempty"`
	// OverrideStatus replaces the response status code if non-zero
	OverrideStatus int `yaml:"override_status,omitempty"`
}

// IsEnabled returns true if any transform rules are configured.
func (t *TransformConfig) IsEnabled() bool {
	if t == nil {
		return false
	}
	return t.Request != nil || t.Response != nil
}
