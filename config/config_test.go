package config

import (
	"os"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "patchwork-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_BasicConfig(t *testing.T) {
	yaml := `
port: 9090
routes:
  - method: GET
    path: /hello
    status: 200
    body: '{"message": "hello"}'
`
	path := writeTempConfig(t, yaml)
	defer os.Remove(path)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if len(cfg.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(cfg.Routes))
	}
	if cfg.Routes[0].Path != "/hello" {
		t.Errorf("expected path /hello, got %s", cfg.Routes[0].Path)
	}
}

func TestLoad_Defaults(t *testing.T) {
	yaml := `
routes:
  - path: /ping
`
	path := writeTempConfig(t, yaml)
	defer os.Remove(path)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Port)
	}
	if cfg.Routes[0].Status != 200 {
		t.Errorf("expected default status 200, got %d", cfg.Routes[0].Status)
	}
	if cfg.Routes[0].Method != "GET" {
		t.Errorf("expected default method GET, got %s", cfg.Routes[0].Method)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
