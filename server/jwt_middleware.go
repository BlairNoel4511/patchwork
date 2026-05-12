package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/patrickward/patchwork/config"
)

// JWTMiddleware validates a Bearer JWT on incoming requests.
// It supports HS256 HMAC verification and optional claim checks.
// RS256 / JWKS is stubbed — extend jwksVerify for production use.
func JWTMiddleware(cfg *config.JWTConfig, next http.Handler) http.Handler {
	if cfg == nil || !cfg.IsEnabled() {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearer(r)
		if token == "" {
			reject(w, cfg)
			return
		}
		claims, ok := verifyHS256(token, cfg.Secret)
		if !ok {
			reject(w, cfg)
			return
		}
		if cfg.Issuer != "" && claims["iss"] != cfg.Issuer {
			reject(w, cfg)
			return
		}
		if cfg.Audience != "" && claims["aud"] != cfg.Audience {
			reject(w, cfg)
			return
		}
		for k, v := range cfg.RequiredClaims {
			if claims[k] != v {
				reject(w, cfg)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(h, "Bearer ")
}

// verifyHS256 validates the signature and decodes the payload claims.
func verifyHS256(token, secret string) (map[string]interface{}, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, false
	}
	sigInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sigInput))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return nil, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, false
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, false
	}
	return claims, true
}

func reject(w http.ResponseWriter, cfg *config.JWTConfig) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(cfg.ResolvedStatusCode())
	w.Write([]byte(cfg.ResolvedBody())) //nolint:errcheck
}
