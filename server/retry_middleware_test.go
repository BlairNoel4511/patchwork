package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/patchwork/config"
)

func TestRetryMiddleware_NoConfig_PassesThrough(t *testing.T) {
	called := 0
	h := RetryMiddleware(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if called != 1 {
		t.Fatalf("expected handler called once, got %d", called)
	}
}

func TestRetryMiddleware_NoRetryOnSuccess(t *testing.T) {
	called := 0
	cfg := &config.RetryConfig{Attempts: 3, BackoffMS: 0}
	h := RetryMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if called != 1 {
		t.Fatalf("expected 1 call on success, got %d", called)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRetryMiddleware_RetriesOn503(t *testing.T) {
	called := 0
	cfg := &config.RetryConfig{Attempts: 3, BackoffMS: 0}
	h := RetryMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if called != 3 {
		t.Fatalf("expected 3 attempts, got %d", called)
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected final status 503, got %d", rec.Code)
	}
}

func TestRetryMiddleware_StopsOnFirstSuccess(t *testing.T) {
	called := 0
	cfg := &config.RetryConfig{Attempts: 5, BackoffMS: 0}
	h := RetryMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		if called < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if called != 3 {
		t.Fatalf("expected 3 calls, got %d", called)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 after recovery, got %d", rec.Code)
	}
}

func TestRetryMiddleware_ExplicitRetryOn(t *testing.T) {
	called := 0
	cfg := &config.RetryConfig{Attempts: 3, BackoffMS: 0, RetryOn: []int{429}}
	h := RetryMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if called != 3 {
		t.Fatalf("expected 3 retries for 429, got %d", called)
	}
}

func TestNextDelay_Multiplier(t *testing.T) {
	cfg := &config.RetryConfig{Multiplier: 2.0, MaxBackoffMS: 0}
	out := nextDelay(100*1e6, cfg) // 100ms in nanoseconds
	if out != 200*1e6 {
		t.Fatalf("expected 200ms, got %v", out)
	}
}

func TestNextDelay_MaxBackoff(t *testing.T) {
	cfg := &config.RetryConfig{Multiplier: 10.0, MaxBackoffMS: 500}
	out := nextDelay(200*1e6, cfg)
	if out != 500*1e6 {
		t.Fatalf("expected capped at 500ms, got %v", out)
	}
}
