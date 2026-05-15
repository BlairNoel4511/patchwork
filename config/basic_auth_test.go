package config_test

import (
	"testing"

	"github.com/user/patchwork/config"
)

func intBasicAuthPtr(v int) *int { return &v }

func TestBasicAuthConfig_IsEnabled_Nil(t *testing.T) {
	var c *config.BasicAuthConfig
	if c.IsEnabled() {
		t.Fatal("expected false for nil config")
	}
}

func TestBasicAuthConfig_IsEnabled_NoUsers(t *testing.T) {
	c := &config.BasicAuthConfig{}
	if c.IsEnabled() {
		t.Fatal("expected false when no users defined")
	}
}

func TestBasicAuthConfig_IsEnabled_WithUsers(t *testing.T) {
	c := &config.BasicAuthConfig{
		Users: map[string]string{"alice": "secret"},
	}
	if !c.IsEnabled() {
		t.Fatal("expected true when users are present")
	}
}

func TestBasicAuthConfig_ResolvedRealm_Default(t *testing.T) {
	c := &config.BasicAuthConfig{}
	if got := c.ResolvedRealm(); got != "Restricted" {
		t.Fatalf("expected 'Restricted', got %q", got)
	}
}

func TestBasicAuthConfig_ResolvedRealm_Custom(t *testing.T) {
	c := &config.BasicAuthConfig{Realm: "MyApp"}
	if got := c.ResolvedRealm(); got != "MyApp" {
		t.Fatalf("expected 'MyApp', got %q", got)
	}
}

func TestBasicAuthConfig_ResolvedStatusCode_Default(t *testing.T) {
	c := &config.BasicAuthConfig{}
	if got := c.ResolvedStatusCode(); got != 401 {
		t.Fatalf("expected 401, got %d", got)
	}
}

func TestBasicAuthConfig_ResolvedStatusCode_Custom(t *testing.T) {
	c := &config.BasicAuthConfig{StatusCode: intBasicAuthPtr(403)}
	if got := c.ResolvedStatusCode(); got != 403 {
		t.Fatalf("expected 403, got %d", got)
	}
}

func TestBasicAuthConfig_ResolvedBody_Default(t *testing.T) {
	c := &config.BasicAuthConfig{}
	if got := c.ResolvedBody(); got != "Unauthorized" {
		t.Fatalf("expected 'Unauthorized', got %q", got)
	}
}

func TestBasicAuthConfig_ResolvedBody_Custom(t *testing.T) {
	c := &config.BasicAuthConfig{Body: "Access Denied"}
	if got := c.ResolvedBody(); got != "Access Denied" {
		t.Fatalf("expected 'Access Denied', got %q", got)
	}
}
