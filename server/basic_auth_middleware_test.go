package server

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/patchwork/config"
)

func makeBasicAuthConfig(user, pass, realm string) *config.BasicAuthConfig {
	return &config.BasicAuthConfig{
		Username: user,
		Password: pass,
		Realm:    realm,
	}
}

func basicAuthHeader(user, pass string) string {
	creds := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	return "Basic " + creds
}

func TestBasicAuthMiddleware_NilConfig_PassesThrough(t *testing.T) {
	h := BasicAuthMiddleware(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestBasicAuthMiddleware_ValidCredentials_Passes(t *testing.T) {
	cfg := makeBasicAuthConfig("alice", "secret", "")
	h := BasicAuthMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", basicAuthHeader("alice", "secret"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestBasicAuthMiddleware_WrongPassword_Returns401(t *testing.T) {
	cfg := makeBasicAuthConfig("alice", "secret", "TestRealm")
	h := BasicAuthMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", basicAuthHeader("alice", "wrong"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if rec.Header().Get("WWW-Authenticate") == "" {
		t.Fatal("expected WWW-Authenticate header")
	}
}

func TestBasicAuthMiddleware_MissingHeader_Returns401(t *testing.T) {
	cfg := makeBasicAuthConfig("alice", "secret", "")
	h := BasicAuthMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestBasicAuthMiddleware_CustomStatusAndBody(t *testing.T) {
	cfg := makeBasicAuthConfig("bob", "pass", "")
	cfg.StatusCode = 403
	cfg.Body = "forbidden"
	h := BasicAuthMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != 403 {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
	if rec.Body.String() != "forbidden" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
