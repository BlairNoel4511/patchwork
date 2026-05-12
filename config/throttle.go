package config

// ThrottleConfig defines request throttling (concurrency limiting) for a route.
type ThrottleConfig struct {
	// MaxConcurrent is the maximum number of in-flight requests allowed at once.
	MaxConcurrent int `yaml:"max_concurrent"`
	// QueueSize is how many requests may wait before being rejected.
	QueueSize int `yaml:"queue_size"`
	// QueueTimeout is how long a queued request will wait before being rejected (e.g. "500ms").
	QueueTimeout string `yaml:"queue_timeout"`
	// StatusCode is returned when the queue is full (default 503).
	StatusCode int `yaml:"status_code"`
	// Body is the response body when throttled.
	Body string `yaml:"body"`
}

// IsEnabled returns true when throttling is configured.
func (t *ThrottleConfig) IsEnabled() bool {
	return t != nil && t.MaxConcurrent > 0
}

// ResolvedStatusCode returns the configured status or 503.
func (t *ThrottleConfig) ResolvedStatusCode() int {
	if t == nil || t.StatusCode == 0 {
		return 503
	}
	return t.StatusCode
}

// ResolvedBody returns the configured body or a default message.
func (t *ThrottleConfig) ResolvedBody() string {
	if t == nil || t.Body == "" {
		return "too many concurrent requests"
	}
	return t.Body
}

// ResolvedQueueSize returns the queue size, defaulting to 0 (no queue).
func (t *ThrottleConfig) ResolvedQueueSize() int {
	if t == nil {
		return 0
	}
	return t.QueueSize
}
