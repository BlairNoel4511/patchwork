package config

import (
	"testing"
)

func boolPtr(b bool) *bool { return &b }

func TestAdminPrefix_Default(t *testing.T) {
	var a *AdminConfig
	if got := a.AdminPrefix(); got != "/__admin" {
		t.Errorf("expected /__admin, got %q", got)
	}
}

func TestAdminPrefix_Custom(t *testing.T) {
	a := &AdminConfig{Prefix: "/_internal"}
	if got := a.AdminPrefix(); got != "/_internal" {
		t.Errorf("expected /_internal, got %q", got)
	}
}

func TestAdminPrefix_TrailingSlashStripped(t *testing.T) {
	// AdminPrefix itself does not strip; callers do — verify raw value returned.
	a := &AdminConfig{Prefix: "/__admin"}
	if got := a.AdminPrefix(); got != "/__admin" {
		t.Errorf("unexpected prefix %q", got)
	}
}

func TestAdminIsEnabled_NilConfig(t *testing.T) {
	var a *AdminConfig
	if !a.IsEnabled() {
		t.Error("nil config should default to enabled")
	}
}

func TestAdminIsEnabled_NilField(t *testing.T) {
	a := &AdminConfig{}
	if !a.IsEnabled() {
		t.Error("nil Enabled field should default to enabled")
	}
}

func TestAdminIsEnabled_ExplicitFalse(t *testing.T) {
	a := &AdminConfig{Enabled: boolPtr(false)}
	if a.IsEnabled() {
		t.Error("expected IsEnabled to return false")
	}
}

func TestAdminIsEnabled_ExplicitTrue(t *testing.T) {
	a := &AdminConfig{Enabled: boolPtr(true)}
	if !a.IsEnabled() {
		t.Error("expected IsEnabled to return true")
	}
}
