package config

import "testing"

func intPtr(v int) *int { return &v }

func TestJWTConfig_IsEnabled_Nil(t *testing.T) {
	var j *JWTConfig
	if j.IsEnabled() {
		t.Fatal("nil config should not be enabled")
	}
}

func TestJWTConfig_IsEnabled_NoSecret(t *testing.T) {
	j := &JWTConfig{}
	if j.IsEnabled() {
		t.Fatal("empty config should not be enabled")
	}
}

func TestJWTConfig_IsEnabled_WithSecret(t *testing.T) {
	j := &JWTConfig{Secret: "mysecret"}
	if !j.IsEnabled() {
		t.Fatal("config with secret should be enabled")
	}
}

func TestJWTConfig_IsEnabled_WithJWKS(t *testing.T) {
	j := &JWTConfig{JWKSURL: "https://example.com/.well-known/jwks.json"}
	if !j.IsEnabled() {
		t.Fatal("config with JWKS URL should be enabled")
	}
}

func TestJWTConfig_ResolvedStatusCode_Default(t *testing.T) {
	j := &JWTConfig{Secret: "s"}
	if got := j.ResolvedStatusCode(); got != 401 {
		t.Fatalf("expected 401, got %d", got)
	}
}

func TestJWTConfig_ResolvedStatusCode_Custom(t *testing.T) {
	j := &JWTConfig{Secret: "s", StatusCode: intPtr(403)}
	if got := j.ResolvedStatusCode(); got != 403 {
		t.Fatalf("expected 403, got %d", got)
	}
}

func TestJWTConfig_ResolvedBody_Default(t *testing.T) {
	j := &JWTConfig{Secret: "s"}
	if got := j.ResolvedBody(); got != `{"error":"unauthorized"}` {
		t.Fatalf("unexpected body: %s", got)
	}
}

func TestJWTConfig_ResolvedBody_Custom(t *testing.T) {
	j := &JWTConfig{Secret: "s", Body: "nope"}
	if got := j.ResolvedBody(); got != "nope" {
		t.Fatalf("unexpected body: %s", got)
	}
}

func TestJWTConfig_NilResolvedStatusCode(t *testing.T) {
	var j *JWTConfig
	if got := j.ResolvedStatusCode(); got != 401 {
		t.Fatalf("expected 401 for nil, got %d", got)
	}
}
