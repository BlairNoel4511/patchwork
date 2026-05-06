package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBodyCacheMiddleware_StoresBody(t *testing.T) {
	payload := `{"hello":"world"}`
	var captured string

	handler := BodyCacheMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Context().Value(bodyKey{})
		if val != nil {
			captured = val.(string)
		}
		w.WriteHeader(http.StatusOK)
	}))

	body := io.NopCloser(strings.NewReader(payload))
	r := httptest.NewRequest(http.MethodPost, "/", body)
	r.ContentLength = int64(len(payload))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if captured != payload {
		t.Errorf("expected body %q, got %q", payload, captured)
	}
}

func TestBodyCacheMiddleware_NilBody(t *testing.T) {
	called := false
	handler := BodyCacheMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if !called {
		t.Error("expected next handler to be called")
	}
	if val := r.Context().Value(bodyKey{}); val != nil {
		t.Error("expected no body in context for nil body request")
	}
}

func TestBodyCacheMiddleware_EmptyContentLength(t *testing.T) {
	handler := BodyCacheMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.ContentLength = 0
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}
