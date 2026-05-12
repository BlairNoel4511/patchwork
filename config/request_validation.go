package config

// RequestValidationConfig defines rules for validating incoming request bodies.
type RequestValidationConfig struct {
	// RequireContentType enforces that the request Content-Type matches one of the listed values.
	RequireContentType []string `yaml:"require_content_type"`
	// MaxBodyBytes sets the maximum allowed request body size in bytes. 0 means unlimited.
	MaxBodyBytes int64 `yaml:"max_body_bytes"`
	// RequireFields lists top-level JSON keys that must be present in the request body.
	RequireFields []string `yaml:"require_fields"`
	// RejectUnknownFields returns 400 if the JSON body contains keys not in AllowedFields.
	RejectUnknownFields bool `yaml:"reject_unknown_fields"`
	// AllowedFields is the set of permitted top-level JSON keys (used with RejectUnknownFields).
	AllowedFields []string `yaml:"allowed_fields"`
	// StatusCode is the HTTP status returned on validation failure. Defaults to 400.
	StatusCode *int `yaml:"status_code"`
	// Body is the response body returned on validation failure.
	Body string `yaml:"body"`
}

// IsEnabled reports whether request validation is configured.
func (r *RequestValidationConfig) IsEnabled() bool {
	if r == nil {
		return false
	}
	return len(r.RequireContentType) > 0 ||
		r.MaxBodyBytes > 0 ||
		len(r.RequireFields) > 0 ||
		r.RejectUnknownFields
}

// ResolvedStatusCode returns the configured status code or 400 as default.
func (r *RequestValidationConfig) ResolvedStatusCode() int {
	if r == nil || r.StatusCode == nil {
		return 400
	}
	return *r.StatusCode
}

// ResolvedBody returns the configured body or a default message.
func (r *RequestValidationConfig) ResolvedBody() string {
	if r == nil || r.Body == "" {
		return "bad request"
	}
	return r.Body
}
