package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zachlatta/patchwork/config"
)

func makeHeaderRoute(reqHeaders, respHeaders map[string]string) config.Route {
	return config.Route{
		Path:            "/test",
		Method:          http.MethodGet,
		RequestHeaders:  reqHeaders,
		ResponseHeaders: respHeaders,
	}
}

func TestHeaderMiddleware_NoHeaders(t *testing.T) {
	route := makeHeaderRoute(nil, nil)
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	h := HeaderMiddleware(route, next)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test", nil))

	if !called {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHeaderMiddleware_InjectsResponseHeaders(t *testing.T) {
	route := makeHeaderRoute(nil, map[string]string{
		"X-Custom-Header": "patchwork",
		"X-Version":       "1",
	})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := HeaderMiddleware(route, next)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test", nil))

	if got := rec.Header().Get("X-Custom-Header"); got != "patchwork" {
		t.Errorf("X-Custom-Header: want 'patchwork', got %q", got)
	}
	if got := rec.Header().Get("X-Version"); got != "1" {
		t.Errorf("X-Version: want '1', got %q", got)
	}
}

func TestHeaderMiddleware_InjectsRequestHeaders(t *testing.T) {
	route := makeHeaderRoute(map[string]string{"X-Forwarded-By": "patchwork"}, nil)

	var capturedHeader string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Forwarded-By")
		w.WriteHeader(http.StatusOK)
	})

	h := HeaderMiddleware(route, next)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test", nil))

	if capturedHeader != "patchwork" {
		t.Errorf("X-Forwarded-By: want 'patchwork', got %q", capturedHeader)
	}
}

func TestHeaderMiddleware_ResponseHeadersOnImplicitWrite(t *testing.T) {
	route := makeHeaderRoute(nil, map[string]string{"X-Implicit": "yes"})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write body without explicit WriteHeader — triggers implicit 200.
		_, _ = w.Write([]byte("hello"))
	})

	h := HeaderMiddleware(route, next)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test", nil))

	if got := rec.Header().Get("X-Implicit"); got != "yes" {
		t.Errorf("X-Implicit: want 'yes', got %q", got)
	}
	if rec.Body.String() != "hello" {
		t.Errorf("unexpected body: %q", rec.Body.String())
	}
}
