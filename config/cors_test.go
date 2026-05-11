package config

import "testing"

func TestCORSConfig_IsEnabled_Nil(t *testing.T) {
	var c *CORSConfig
	if c.IsEnabled() {
		t.Error("expected nil CORSConfig to not be enabled")
	}
}

func TestCORSConfig_IsEnabled_NoOrigins(t *testing.T) {
	c := &CORSConfig{}
	if c.IsEnabled() {
		t.Error("expected CORSConfig with no origins to not be enabled")
	}
}

func TestCORSConfig_IsEnabled_WithOrigins(t *testing.T) {
	c := &CORSConfig{AllowedOrigins: []string{"https://example.com"}}
	if !c.IsEnabled() {
		t.Error("expected CORSConfig with origins to be enabled")
	}
}

func TestCORSConfig_AllowedOriginsValue_Wildcard(t *testing.T) {
	c := &CORSConfig{AllowedOrigins: []string{"https://a.com", "*"}}
	if got := c.AllowedOriginsValue(); got != "*" {
		t.Errorf("expected *, got %q", got)
	}
}

func TestCORSConfig_AllowedOriginsValue_Multiple(t *testing.T) {
	c := &CORSConfig{AllowedOrigins: []string{"https://a.com", "https://b.com"}}
	want := "https://a.com, https://b.com"
	if got := c.AllowedOriginsValue(); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestCORSConfig_AllowedOriginsValue_Nil(t *testing.T) {
	var c *CORSConfig
	if got := c.AllowedOriginsValue(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestCORSConfig_AllowedMethodsValue_Default(t *testing.T) {
	c := &CORSConfig{}
	want := "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	if got := c.AllowedMethodsValue(); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestCORSConfig_AllowedMethodsValue_Custom(t *testing.T) {
	c := &CORSConfig{AllowedMethods: []string{"GET", "POST"}}
	want := "GET, POST"
	if got := c.AllowedMethodsValue(); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestCORSConfig_AllowedHeadersValue_Default(t *testing.T) {
	c := &CORSConfig{}
	want := "Content-Type, Authorization"
	if got := c.AllowedHeadersValue(); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestCORSConfig_AllowedHeadersValue_Custom(t *testing.T) {
	c := &CORSConfig{AllowedHeaders: []string{"X-Api-Key", "Accept"}}
	want := "X-Api-Key, Accept"
	if got := c.AllowedHeadersValue(); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
