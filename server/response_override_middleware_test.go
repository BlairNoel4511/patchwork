package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/patrickward/patchwork/config"
)

func resetOverrideStore() {
	globalOverrideStore.reset()
}

func TestResponseOverrideMiddleware_NoOverride_PassesThrough(t *testing.T) {
	resetOverrideStore()
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	h := ResponseOverrideMiddleware("/api/foo", next)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/foo", nil))
	if !called {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestResponseOverrideMiddleware_ActiveOverride_ShortCircuits(t *testing.T) {
	resetOverrideStore()
	globalOverrideStore.set("/api/foo", &config.ResponseOverride{
		Status: http.StatusServiceUnavailable,
		Body:   "override body",
	})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler must not be called")
	})
	h := ResponseOverrideMiddleware("/api/foo", next)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/foo", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "override body") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestResponseOverrideMiddleware_InjectsHeaders(t *testing.T) {
	resetOverrideStore()
	globalOverrideStore.set("/api/bar", &config.ResponseOverride{
		Status:  http.StatusOK,
		Headers: map[string]string{"X-Overridden": "true"},
	})
	h := ResponseOverrideMiddleware("/api/bar", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach next")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/bar", nil))
	if rec.Header().Get("X-Overridden") != "true" {
		t.Fatalf("expected X-Overridden header, got %q", rec.Header().Get("X-Overridden"))
	}
}

func TestResponseOverrideAdminHandler_SetAndClear(t *testing.T) {
	resetOverrideStore()
	h := ResponseOverrideAdminHandler()

	// Set an override via PUT
	body := strings.NewReader(`{"status":503,"body":"down"}`)
	req := httptest.NewRequest(http.MethodPut, "/?route=/api/test", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on PUT, got %d", rec.Code)
	}
	if o, ok := globalOverrideStore.get("/api/test"); !ok || o.Status != 503 {
		t.Fatal("override not stored correctly")
	}

	// Clear the specific override via DELETE
	req2 := httptest.NewRequest(http.MethodDelete, "/?route=/api/test", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on DELETE, got %d", rec2.Code)
	}
	if _, ok := globalOverrideStore.get("/api/test"); ok {
		t.Fatal("override should have been removed")
	}
}
