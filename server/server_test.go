package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/patchwork/config"
)

func makeTestConfig(routes []config.Route) *config.Config {
	return &config.Config{
		Server: config.Server{Port: 8080, MaxLogEntries: 100},
		Routes: routes,
	}
}

func TestNew_RoutesAreRegistered(t *testing.T) {
	cfg := makeTestConfig([]config.Route{
		{Method: "GET", Path: "/hello", Response: config.Response{Status: 200, Body: "hi"}},
	})
	h := New(cfg)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/hello", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNew_CORSHeaderPresent(t *testing.T) {
	cfg := makeTestConfig([]config.Route{
		{Method: "GET", Path: "/cors", Response: config.Response{Status: 200}},
	})
	h := New(cfg)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/cors", nil))
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("expected CORS header")
	}
}

func TestNew_UnknownRouteReturns404(t *testing.T) {
	cfg := makeTestConfig(nil)
	h := New(cfg)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/nope", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestNew_RateLimitedRouteReturns429(t *testing.T) {
	// Reset store
	globalRateLimitStore.mu.Lock()
	globalRateLimitStore.buckets = make(map[string]*tokenBucket)
	globalRateLimitStore.mu.Unlock()

	cfg := makeTestConfig([]config.Route{
		{
			Method:   "GET",
			Path:     "/limited",
			Response: config.Response{Status: 200, Body: "ok"},
			RateLimit: &config.RateLimit{
				RequestsPerSecond: 0.001,
				Burst:             1,
			},
		},
	})
	h := New(cfg)
	// First request passes
	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, httptest.NewRequest("GET", "/limited", nil))
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec1.Code)
	}
	// Second request is rate limited
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest("GET", "/limited", nil))
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec2.Code)
	}
}
