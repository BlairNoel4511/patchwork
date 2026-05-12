package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/patchwork/config"
)

func TestTransformMiddleware_NilConfig_PassesThrough(t *testing.T) {
	called := false
	h := TransformMiddleware(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !called {
		t.Fatal("expected next handler to be called")
	}
}

func TestTransformMiddleware_StripPathPrefix(t *testing.T) {
	var gotPath string
	cfg := &config.TransformConfig{Request: &config.RequestTransform{StripPathPrefix: "/api/v1"}}
	h := TransformMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/users", nil))
	if gotPath != "/users" {
		t.Fatalf("expected /users, got %s", gotPath)
	}
}

func TestTransformMiddleware_AddPathPrefix(t *testing.T) {
	var gotPath string
	cfg := &config.TransformConfig{Request: &config.RequestTransform{AddPathPrefix: "/v2"}}
	h := TransformMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/users", nil))
	if gotPath != "/v2/users" {
		t.Fatalf("expected /v2/users, got %s", gotPath)
	}
}

func TestTransformMiddleware_SetRequestHeaders(t *testing.T) {
	var gotHeader string
	cfg := &config.TransformConfig{Request: &config.RequestTransform{
		SetHeaders: map[string]string{"X-Tenant": "acme"},
	}}
	h := TransformMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Tenant")
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if gotHeader != "acme" {
		t.Fatalf("expected acme, got %s", gotHeader)
	}
}

func TestTransformMiddleware_RemoveRequestHeaders(t *testing.T) {
	var gotHeader string
	cfg := &config.TransformConfig{Request: &config.RequestTransform{
		RemoveHeaders: []string{"Authorization"},
	}}
	h := TransformMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if gotHeader != "" {
		t.Fatalf("expected empty Authorization, got %s", gotHeader)
	}
}

func TestTransformMiddleware_OverrideResponseStatus(t *testing.T) {
	cfg := &config.TransformConfig{Response: &config.ResponseTransform{OverrideStatus: http.StatusAccepted}}
	h := TransformMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
}

func TestTransformMiddleware_SetResponseHeaders(t *testing.T) {
	cfg := &config.TransformConfig{Response: &config.ResponseTransform{
		SetHeaders: map[string]string{"X-Powered-By": "patchwork"},
	}}
	h := TransformMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("X-Powered-By") != "patchwork" {
		t.Fatalf("expected patchwork header, got %s", rec.Header().Get("X-Powered-By"))
	}
}
