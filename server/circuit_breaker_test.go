package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/damiensedgwick/patchwork/config"
)

func makeCBConfig(threshold, openDurationMs, statusCode int, body string) config.CircuitBreakerConfig {
	return config.CircuitBreakerConfig{
		Enabled:      true,
		Threshold:    threshold,
		OpenDuration: openDurationMs,
		StatusCode:   statusCode,
		Body:         body,
	}
}

func TestCircuitBreaker_ClosedByDefault(t *testing.T) {
	cb := newCircuitBreaker(makeCBConfig(3, 1000, 503, ""))
	if !cb.allow() {
		t.Fatal("expected circuit to be closed initially")
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := newCircuitBreaker(makeCBConfig(3, 60000, 503, ""))
	cb.recordFailure()
	cb.recordFailure()
	cb.recordFailure()
	if cb.allow() {
		t.Fatal("expected circuit to be open after threshold failures")
	}
}

func TestCircuitBreaker_ResetsOnSuccess(t *testing.T) {
	cb := newCircuitBreaker(makeCBConfig(3, 60000, 503, ""))
	cb.recordFailure()
	cb.recordFailure()
	cb.recordSuccess()
	if !cb.allow() {
		t.Fatal("expected circuit to remain closed after success resets failures")
	}
}

func TestCircuitBreakerMiddleware_AllowsNormalRequests(t *testing.T) {
	cfg := makeCBConfig(3, 60000, 503, "open")
	cb := newCircuitBreaker(cfg)
	handler := CircuitBreakerMiddleware(cfg, cb)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCircuitBreakerMiddleware_RejectsWhenOpen(t *testing.T) {
	cfg := makeCBConfig(2, 60000, 503, "circuit open")
	cb := newCircuitBreaker(cfg)
	// Trip the circuit by recording failures directly.
	cb.recordFailure()
	cb.recordFailure()

	handler := CircuitBreakerMiddleware(cfg, cb)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != 503 {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	if rec.Body.String() != "circuit open" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestCircuitBreakerMiddleware_TripsOn5xx(t *testing.T) {
	cfg := makeCBConfig(2, 60000, 503, "")
	cb := newCircuitBreaker(cfg)
	handler := CircuitBreakerMiddleware(cfg, cb)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	if cb.state != stateOpen {
		t.Fatal("expected circuit to be open after consecutive 5xx responses")
	}
}
