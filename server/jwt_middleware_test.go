package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/patchwork/config"
)

func makeJWT(secret string, claims map[string]interface{}) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, _ := json.Marshal(claims)
	enc := base64.RawURLEncoding.EncodeToString(payload)
	sigInput := header + "." + enc
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sigInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s.%s.%s", header, enc, sig)
}

var jwtOKHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestJWTMiddleware_NilConfig_PassesThrough(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	JWTMiddleware(nil, jwtOKHandler).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestJWTMiddleware_MissingToken_Returns401(t *testing.T) {
	cfg := &config.JWTConfig{Secret: "secret"}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	JWTMiddleware(cfg, jwtOKHandler).ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWTMiddleware_ValidToken_Passes(t *testing.T) {
	cfg := &config.JWTConfig{Secret: "topsecret"}
	token := makeJWT("topsecret", map[string]interface{}{"sub": "user1"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	JWTMiddleware(cfg, jwtOKHandler).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestJWTMiddleware_WrongSecret_Returns401(t *testing.T) {
	cfg := &config.JWTConfig{Secret: "correct"}
	token := makeJWT("wrong", map[string]interface{}{"sub": "user1"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	JWTMiddleware(cfg, jwtOKHandler).ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWTMiddleware_IssuerMismatch_Returns401(t *testing.T) {
	cfg := &config.JWTConfig{Secret: "s", Issuer: "expected-issuer"}
	token := makeJWT("s", map[string]interface{}{"iss": "other"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	JWTMiddleware(cfg, jwtOKHandler).ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWTMiddleware_RequiredClaimsMismatch_Returns401(t *testing.T) {
	cfg := &config.JWTConfig{Secret: "s", RequiredClaims: map[string]string{"role": "admin"}}
	token := makeJWT("s", map[string]interface{}{"role": "user"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	JWTMiddleware(cfg, jwtOKHandler).ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWTMiddleware_RequiredClaimsMatch_Passes(t *testing.T) {
	cfg := &config.JWTConfig{Secret: "s", RequiredClaims: map[string]string{"role": "admin"}}
	token := makeJWT("s", map[string]interface{}{"role": "admin"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	JWTMiddleware(cfg, jwtOKHandler).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
