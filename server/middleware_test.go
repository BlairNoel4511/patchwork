package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingMiddleware_CapturesStatus(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	wrapped := LoggingMiddleware(handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}

func TestCORSMiddleware_AddsHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := CORSMiddleware(handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	wrapped.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected Access-Control-Allow-Origin *, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Access-Control-Allow-Methods to be set")
	}
}

func TestCORSMiddleware_PreflightReturns204(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := CORSMiddleware(handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS preflight, got %d", rec.Code)
	}
}

func TestChain_AppliesMiddlewareInOrder(t *testing.T) {
	var order []string

	makeMiddleware := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	chained := Chain(base, makeMiddleware("first"), makeMiddleware("second"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	chained.ServeHTTP(rec, req)

	expected := []string{"first", "second", "handler"}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("expected order[%d]=%q, got %q", i, v, order[i])
		}
	}
}
