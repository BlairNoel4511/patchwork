package config

import "time"

// TimeoutConfig defines per-route request timeout settings.
type TimeoutConfig struct {
	// Duration is a human-readable duration string, e.g. "5s", "500ms".
	Duration string `yaml:"duration"`
	// StatusCode is returned when the timeout is exceeded. Defaults to 504.
	StatusCode int `yaml:"status_code"`
	// Body is the response body returned on timeout.
	Body string `yaml:"body"`
}

// Timeout returns the parsed duration, or 0 if unset / unparseable.
func (t *TimeoutConfig) Timeout() time.Duration {
	if t == nil || t.Duration == "" {
		return 0
	}
	d, err := time.ParseDuration(t.Duration)
	if err != nil {
		return 0
	}
	return d
}

// ResponseStatus returns the configured status code, defaulting to 504.
func (t *TimeoutConfig) ResponseStatus() int {
	if t == nil || t.StatusCode == 0 {
		return 504
	}
	return t.StatusCode
}

// ResponseBody returns the configured body, defaulting to a plain message.
func (t *TimeoutConfig) ResponseBody() string {
	if t == nil || t.Body == "" {
		return "gateway timeout"
	}
	return t.Body
}
