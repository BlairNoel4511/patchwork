package config

import "fmt"

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

// Validate checks that the Webhook has the minimum required fields and that
// its values are sensible. It returns an error describing the first problem
// found, or nil if the configuration is valid.
func (w *Webhook) Validate() error {
	if w.URL == "" {
		return fmt.Errorf("webhook: url is required")
	}
	if w.DelayMs < 0 {
		return fmt.Errorf("webhook: delay_ms must be non-negative, got %d", w.DelayMs)
	}
	return nil
}
