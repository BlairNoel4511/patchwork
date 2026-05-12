package config

import "testing"

func TestThrottleConfig_IsEnabled_Nil(t *testing.T) {
	var tc *ThrottleConfig
	if tc.IsEnabled() {
		t.Fatal("expected nil config to not be enabled")
	}
}

func TestThrottleConfig_IsEnabled_ZeroConcurrent(t *testing.T) {
	tc := &ThrottleConfig{MaxConcurrent: 0}
	if tc.IsEnabled() {
		t.Fatal("expected MaxConcurrent=0 to not be enabled")
	}
}

func TestThrottleConfig_IsEnabled_Positive(t *testing.T) {
	tc := &ThrottleConfig{MaxConcurrent: 5}
	if !tc.IsEnabled() {
		t.Fatal("expected MaxConcurrent=5 to be enabled")
	}
}

func TestThrottleConfig_ResolvedStatusCode_Default(t *testing.T) {
	tc := &ThrottleConfig{}
	if got := tc.ResolvedStatusCode(); got != 503 {
		t.Fatalf("expected 503, got %d", got)
	}
}

func TestThrottleConfig_ResolvedStatusCode_Custom(t *testing.T) {
	tc := &ThrottleConfig{StatusCode: 429}
	if got := tc.ResolvedStatusCode(); got != 429 {
		t.Fatalf("expected 429, got %d", got)
	}
}

func TestThrottleConfig_ResolvedBody_Default(t *testing.T) {
	tc := &ThrottleConfig{}
	if got := tc.ResolvedBody(); got != "too many concurrent requests" {
		t.Fatalf("unexpected body: %q", got)
	}
}

func TestThrottleConfig_ResolvedBody_Custom(t *testing.T) {
	tc := &ThrottleConfig{Body: "overloaded"}
	if got := tc.ResolvedBody(); got != "overloaded" {
		t.Fatalf("unexpected body: %q", got)
	}
}

func TestThrottleConfig_ResolvedQueueSize_Nil(t *testing.T) {
	var tc *ThrottleConfig
	if got := tc.ResolvedQueueSize(); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestThrottleConfig_ResolvedQueueSize_Set(t *testing.T) {
	tc := &ThrottleConfig{QueueSize: 10}
	if got := tc.ResolvedQueueSize(); got != 10 {
		t.Fatalf("expected 10, got %d", got)
	}
}
