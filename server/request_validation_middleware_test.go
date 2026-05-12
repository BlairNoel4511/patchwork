package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/patchwork/config"
)

func intValPtr(v int) *int { return &v }

func okValidationHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequestValidationMiddleware_NilConfig_PassesThrough(t *testing.T) {
	h := RequestValidationMiddleware(nil, okValidationHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequestValidationMiddleware_ContentType_Passes(t *testing.T) {
	cfg := &config.RequestValidationConfig{RequireContentType: []string{"application/json"}}
	h := RequestValidationMiddleware(cfg, okValidationHandler())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequestValidationMiddleware_ContentType_Rejects(t *testing.T) {
	cfg := &config.RequestValidationConfig{RequireContentType: []string{"application/json"}}
	h := RequestValidationMiddleware(cfg, okValidationHandler())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`data`))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRequestValidationMiddleware_RequireFields_Passes(t *testing.T) {
	cfg := &config.RequestValidationConfig{RequireFields: []string{"name", "email"}}
	h := RequestValidationMiddleware(cfg, okValidationHandler())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Alice","email":"a@b.com"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequestValidationMiddleware_RequireFields_Rejects(t *testing.T) {
	cfg := &config.RequestValidationConfig{RequireFields: []string{"name", "email"}}
	h := RequestValidationMiddleware(cfg, okValidationHandler())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Alice"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRequestValidationMiddleware_MaxBodyBytes_Rejects(t *testing.T) {
	cfg := &config.RequestValidationConfig{MaxBodyBytes: 5}
	h := RequestValidationMiddleware(cfg, okValidationHandler())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"key":"value"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRequestValidationMiddleware_RejectUnknownFields(t *testing.T) {
	cfg := &config.RequestValidationConfig{
		RejectUnknownFields: true,
		AllowedFields:       []string{"name"},
	}
	h := RequestValidationMiddleware(cfg, okValidationHandler())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Alice","extra":"bad"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRequestValidationMiddleware_CustomStatusAndBody(t *testing.T) {
	code := 422
	cfg := &config.RequestValidationConfig{
		RequireFields: []string{"id"},
		StatusCode:    &code,
		Body:          "unprocessable",
	}
	h := RequestValidationMiddleware(cfg, okValidationHandler())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 422 {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "unprocessable") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
