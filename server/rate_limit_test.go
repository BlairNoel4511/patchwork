package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/patchwork/config"
)

func makeRateLimitRoute(rps float64, burst int) config.Route {
	return config.Route{
		Method: "GET",
		Path:   "/rl-test",
		RateLimit: &config.RateLimit{
			RequestsPerSecond: rps,
			Burst:             burst,
		},
	}
}

func TestRateLimitMiddleware_NoConfig(t *testing.T) {
	route := config.Route{Method: "GET", Path: "/no-rl"}
	h := RateLimitMiddleware(route)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/no-rl", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRateLimitMiddleware_AllowsWithinBurst(t *testing.T) {
	// Reset store to avoid cross-test pollution
	globalRateLimitStore.mu.Lock()
	globalRateLimitStore.buckets = make(map[string]*tokenBucket)
	globalRateLimitStore.mu.Unlock()

	route := makeRateLimitRoute(100, 3)
	route.Path = "/rl-burst"
	h := RateLimitMiddleware(route)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "/rl-burst", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, rec.Code)
		}
	}
}

func TestRateLimitMiddleware_BlocksWhenExceeded(t *testing.T) {
	globalRateLimitStore.mu.Lock()
	globalRateLimitStore.buckets = make(map[string]*tokenBucket)
	globalRateLimitStore.mu.Unlock()

	route := makeRateLimitRoute(0.001, 1)
	route.Path = "/rl-block"
	h := RateLimitMiddleware(route)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	// First request consumes the single token
	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, httptest.NewRequest("GET", "/rl-block", nil))
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", rec1.Code)
	}
	// Second request should be rate-limited
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest("GET", "/rl-block", nil))
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: expected 429, got %d", rec2.Code)
	}
}

func TestRateLimitMiddleware_CustomStatusAndBody(t *testing.T) {
	globalRateLimitStore.mu.Lock()
	globalRateLimitStore.buckets = make(map[string]*tokenBucket)
	globalRateLimitStore.mu.Unlock()

	route := config.Route{
		Method: "GET",
		Path:   "/rl-custom",
		RateLimit: &config.RateLimit{
			RequestsPerSecond: 0.001,
			Burst:             1,
			StatusCode:        503,
			Body:              "slow down",
		},
	}
	h := RateLimitMiddleware(route)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	// Exhaust the bucket
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/rl-custom", nil))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/rl-custom", nil))
	if rec.Code != 503 {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "slow down\n" {
		t.Fatalf("unexpected body: %q", body)
	}
}
