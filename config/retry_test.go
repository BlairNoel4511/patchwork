package config

import "testing"

func TestRetryConfig_NilSafe(t *testing.T) {
	var r *RetryConfig
	if err := r.Validate(); err != nil {
		t.Fatalf("expected nil error for nil config, got %v", err)
	}
	if r.IsEnabled() {
		t.Fatal("expected IsEnabled to return false for nil config")
	}
	if r.ShouldRetry(503) {
		t.Fatal("expected ShouldRetry to return false for nil config")
	}
}

func TestRetryConfig_Validate_Valid(t *testing.T) {
	r := &RetryConfig{Attempts: 3, BackoffMS: 100, Multiplier: 2.0, MaxBackoffMS: 1000}
	if err := r.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetryConfig_Validate_InvalidAttempts(t *testing.T) {
	r := &RetryConfig{Attempts: 0}
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for attempts < 1")
	}
}

func TestRetryConfig_Validate_NegativeBackoff(t *testing.T) {
	r := &RetryConfig{Attempts: 2, BackoffMS: -1}
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for negative backoff_ms")
	}
}

func TestRetryConfig_IsEnabled(t *testing.T) {
	if (&RetryConfig{Attempts: 1}).IsEnabled() {
		t.Fatal("attempts=1 should not be considered enabled (no retry)")
	}
	if !(&RetryConfig{Attempts: 3}).IsEnabled() {
		t.Fatal("attempts=3 should be enabled")
	}
}

func TestRetryConfig_ShouldRetry_DefaultOn5xx(t *testing.T) {
	r := &RetryConfig{Attempts: 3}
	if !r.ShouldRetry(503) {
		t.Fatal("expected ShouldRetry true for 503 with empty RetryOn")
	}
	if r.ShouldRetry(404) {
		t.Fatal("expected ShouldRetry false for 404 with empty RetryOn")
	}
}

func TestRetryConfig_ShouldRetry_ExplicitCodes(t *testing.T) {
	r := &RetryConfig{Attempts: 3, RetryOn: []int{429, 503}}
	if !r.ShouldRetry(429) {
		t.Fatal("expected true for 429")
	}
	if r.ShouldRetry(500) {
		t.Fatal("expected false for 500 when not in RetryOn list")
	}
}
