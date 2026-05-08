package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/patchwork/config"
)

func makeAdminConfig(enabled bool, prefix string, readOnly bool) *config.Config {
	en := enabled
	return &config.Config{
		Admin: &config.AdminConfig{
			Enabled:  &en,
			Prefix:   prefix,
			ReadOnly: readOnly,
		},
		Routes: []config.Route{
			{Path: "/foo"},
			{Path: "/bar"},
		},
	}
}

func TestAdminHealth(t *testing.T) {
	mux := http.NewServeMux()
	AdminRouter(mux, makeAdminConfig(true, "/__admin", false), NewRequestLog(50), &ScenarioStore{})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/__admin/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %q", body["status"])
	}
}

func TestAdminRoutes_ListsPaths(t *testing.T) {
	mux := http.NewServeMux()
	AdminRouter(mux, makeAdminConfig(true, "/__admin", false), NewRequestLog(50), &ScenarioStore{})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/__admin/routes", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&body)
	routes, ok := body["routes"].([]any)
	if !ok || len(routes) != 2 {
		t.Errorf("expected 2 routes, got %v", body["routes"])
	}
}

func TestAdminRequests_DeleteClearsLog(t *testing.T) {
	log := NewRequestLog(50)
	log.Add(LogEntry{Path: "/test"})
	mux := http.NewServeMux()
	AdminRouter(mux, makeAdminConfig(true, "/__admin", false), log, &ScenarioStore{})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/__admin/requests", nil))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if len(log.All()) != 0 {
		t.Error("expected log to be empty after DELETE")
	}
}

func TestAdminRequests_ReadOnlyBlocks(t *testing.T) {
	mux := http.NewServeMux()
	AdminRouter(mux, makeAdminConfig(true, "/__admin", true), NewRequestLog(50), &ScenarioStore{})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/__admin/requests", nil))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAdmin_DisabledRegistersNoRoutes(t *testing.T) {
	mux := http.NewServeMux()
	AdminRouter(mux, makeAdminConfig(false, "/__admin", false), NewRequestLog(50), &ScenarioStore{})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/__admin/health", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when admin disabled, got %d", rec.Code)
	}
}
