package config

import (
	"testing"
	"time"
)

func TestTimeoutConfig_NilSafe(t *testing.T) {
	var tc *TimeoutConfig
	if got := tc.Timeout(); got != 0 {
		t.Errorf("expected 0, got %v", got)
	}
	if got := tc.ResponseStatus(); got != 504 {
		t.Errorf("expected 504, got %d", got)
	}
	if got := tc.ResponseBody(); got != "gateway timeout" {
		t.Errorf("expected default body, got %q", got)
	}
}

func TestTimeoutConfig_ParsesDuration(t *testing.T) {
	tc := &TimeoutConfig{Duration: "2s"}
	if got := tc.Timeout(); got != 2*time.Second {
		t.Errorf("expected 2s, got %v", got)
	}
}

func TestTimeoutConfig_InvalidDuration(t *testing.T) {
	tc := &TimeoutConfig{Duration: "not-a-duration"}
	if got := tc.Timeout(); got != 0 {
		t.Errorf("expected 0 for invalid duration, got %v", got)
	}
}

func TestTimeoutConfig_CustomStatusAndBody(t *testing.T) {
	tc := &TimeoutConfig{
		Duration:   "100ms",
		StatusCode: 408,
		Body:       "request timed out",
	}
	if got := tc.ResponseStatus(); got != 408 {
		t.Errorf("expected 408, got %d", got)
	}
	if got := tc.ResponseBody(); got != "request timed out" {
		t.Errorf("expected custom body, got %q", got)
	}
}
