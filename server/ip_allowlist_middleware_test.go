package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/patchwork/config"
)

func makeAllowlistConfig(allow []string, trustProxy bool) *config.IPAllowlistConfig {
	return &config.IPAllowlistConfig{
		Allow:      allow,
		TrustProxy: trustProxy,
	}
}

var allowPassHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestIPAllowlistMiddleware_NilConfig_PassesThrough(t *testing.T) {
	h := IPAllowlistMiddleware(nil, allowPassHandler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:9000"
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestIPAllowlistMiddleware_AllowedExactIP(t *testing.T) {
	cfg := makeAllowlistConfig([]string{"192.168.1.10"}, false)
	h := IPAllowlistMiddleware(cfg, allowPassHandler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.10:5000"
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestIPAllowlistMiddleware_BlockedIP(t *testing.T) {
	cfg := makeAllowlistConfig([]string{"10.0.0.1"}, false)
	h := IPAllowlistMiddleware(cfg, allowPassHandler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.99.1:1234"
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestIPAllowlistMiddleware_AllowedCIDR(t *testing.T) {
	cfg := makeAllowlistConfig([]string{"10.0.0.0/8"}, false)
	h := IPAllowlistMiddleware(cfg, allowPassHandler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.42.0.5:8080"
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestIPAllowlistMiddleware_TrustProxy_XForwardedFor(t *testing.T) {
	cfg := makeAllowlistConfig([]string{"203.0.113.5"}, true)
	h := IPAllowlistMiddleware(cfg, allowPassHandler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:9000" // internal proxy
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestIPAllowlistMiddleware_CustomDenyStatus(t *testing.T) {
	status := 401
	cfg := &config.IPAllowlistConfig{
		Allow:      []string{"127.0.0.1"},
		DenyStatus: &status,
		DenyBody:   "nope",
	}
	h := IPAllowlistMiddleware(cfg, allowPassHandler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "8.8.8.8:53"
	h.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
