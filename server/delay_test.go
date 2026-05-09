package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDelayMiddleware_NoDelay(t *testing.T) {
	handler := DelayMiddleware(0)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	start := time.Now()
	handler.ServeHTTP(rr, req)
	elapsed := time.Since(start)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if elapsed > 50*time.Millisecond {
		t.Errorf("expected no delay, but took %v", elapsed)
	}
}

func TestDelayMiddleware_WithDelay(t *testing.T) {
	const delayMs = 100
	handler := DelayMiddleware(delayMs)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	start := time.Now()
	handler.ServeHTTP(rr, req)
	elapsed := time.Since(start)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if elapsed < time.Duration(delayMs)*time.Millisecond {
		t.Errorf("expected delay of at least %dms, but took %v", delayMs, elapsed)
	}
}

func TestDelayMiddleware_NegativeDelay(t *testing.T) {
	handler := DelayMiddleware(-10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	start := time.Now()
	handler.ServeHTTP(rr, req)
	elapsed := time.Since(start)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if elapsed > 50*time.Millisecond {
		t.Errorf("negative delay should not sleep, but took %v", elapsed)
	}
}

func TestDelayMiddleware_PassesRequestThrough(t *testing.T) {
	var capturedPath string
	handler := DelayMiddleware(0)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	handler.ServeHTTP(rr, req)

	if capturedPath != "/test-path" {
		t.Errorf("expected request path to be passed through, got %q", capturedPath)
	}
}
