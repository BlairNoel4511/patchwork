package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/patchwork/config"
)

func makeChaosRoute(rate float64, status int, body string) config.Route {
	return config.Route{
		Chaos: &config.ChaosConfig{
			ErrorRate:  rate,
			StatusCode: status,
			Body:       body,
		},
	}
}

func TestChaosMiddleware_NoChaosConfig(t *testing.T) {
	route := config.Route{}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := ChaosMiddleware(route)(next)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestChaosMiddleware_ZeroErrorRate(t *testing.T) {
	route := makeChaosRoute(0.0, 500, "boom")
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := ChaosMiddleware(route)(next)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestChaosMiddleware_AlwaysFails(t *testing.T) {
	route := makeChaosRoute(1.0, http.StatusServiceUnavailable, "chaos error")
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := ChaosMiddleware(route)(next)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
	if rr.Body.String() != "chaos error" {
		t.Errorf("unexpected body: %s", rr.Body.String())
	}
	if rr.Header().Get("X-Patchwork-Chaos") != "true" {
		t.Error("expected X-Patchwork-Chaos header to be set")
	}
}

func TestChaosMiddleware_DefaultStatusAndBody(t *testing.T) {
	route := makeChaosRoute(1.0, 0, "")
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := ChaosMiddleware(route)(next)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
	if rr.Body.String() != http.StatusText(http.StatusInternalServerError) {
		t.Errorf("unexpected body: %s", rr.Body.String())
	}
}
