package config

// Webhook describes an outbound HTTP call that patchwork makes after serving
// a matched route. This allows tests to simulate callback / event-driven
// flows without an external orchestration layer.
type Webhook struct {
	// Method is the HTTP verb used for the outbound request (default: POST).
	Method string `yaml:"method"`

	// URL is the target endpoint. Required.
	URL string `yaml:"url"`

	// Headers are optional key/value pairs added to the outbound request.
	Headers map[string]string `yaml:"headers,omitempty"`

	// Body is an optional request body sent with the outbound request.
	Body string `yaml:"body,omitempty"`

	// DelayMs is the number of milliseconds to wait before firing the
	// webhook. The delay runs in a goroutine so it never blocks the
	// mock response.
	DelayMs int `yaml:"delay_ms,omitempty"`
}
