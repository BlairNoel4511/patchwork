package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/patchwork/config"
)

func boolRequestIDPtr(b bool) *bool { return &b }

func makeRequestIDConfig(enabled *bool, header string, forceNew bool) *config.RequestIDConfig {
	return &config.RequestIDConfig{
		Enabled:  enabled,
		Header:   header,
		ForceNew: forceNew,
	}
}

func TestRequestIDMiddleware_NilConfig_PassesThrough(t *testing.T) {
	var cfg *config.RequestIDConfig
	h := RequestIDMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Request-Id") != "" {
		t.Fatal("expected no request-id header when config is nil")
	}
}

func TestRequestIDMiddleware_SetsDefaultHeader(t *testing.T) {
	cfg := makeRequestIDConfig(boolRequestIDPtr(true), "", false)
	h := RequestIDMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("X-Request-Id") == "" {
		t.Fatal("expected X-Request-Id header to be set")
	}
}

func TestRequestIDMiddleware_PreservesIncomingID(t *testing.T) {
	cfg := makeRequestIDConfig(boolRequestIDPtr(true), "X-Request-Id", false)
	var captured string
	h := RequestIDMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", "my-stable-id")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if captured != "my-stable-id" {
		t.Fatalf("expected preserved id 'my-stable-id', got %q", captured)
	}
	if rec.Header().Get("X-Request-Id") != "my-stable-id" {
		t.Fatal("expected response header to echo incoming id")
	}
}

func TestRequestIDMiddleware_ForceNew_OverridesIncoming(t *testing.T) {
	cfg := makeRequestIDConfig(boolRequestIDPtr(true), "X-Request-Id", true)
	var captured string
	h := RequestIDMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", "old-id")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if captured == "old-id" {
		t.Fatal("expected a new ID to be generated, but old-id was kept")
	}
	if captured == "" {
		t.Fatal("expected a non-empty generated ID")
	}
}

func TestRequestIDMiddleware_CustomHeader(t *testing.T) {
	cfg := makeRequestIDConfig(boolRequestIDPtr(true), "X-Trace-Id", false)
	h := RequestIDMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("X-Trace-Id") == "" {
		t.Fatal("expected X-Trace-Id header to be set")
	}
	if rec.Header().Get("X-Request-Id") != "" {
		t.Fatal("expected default header to be absent when custom header is configured")
	}
}
