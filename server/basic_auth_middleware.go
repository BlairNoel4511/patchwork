package server

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/patrickward/patchwork/config"
)

// BasicAuthMiddleware enforces HTTP Basic Authentication on a route when
// the route's BasicAuth config is enabled.
func BasicAuthMiddleware(cfg *config.BasicAuthConfig, next http.Handler) http.Handler {
	if cfg == nil || !cfg.IsEnabled() {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			rejectBasicAuth(w, cfg)
			return
		}

		expectedUser := []byte(cfg.Username)
		expectedPass := []byte(cfg.Password)
		givenUser := []byte(username)
		givenPass := []byte(password)

		// Use constant-time comparison to prevent timing attacks.
		userMatch := subtle.ConstantTimeCompare(givenUser, expectedUser) == 1
		passMatch := subtle.ConstantTimeCompare(givenPass, expectedPass) == 1

		if !userMatch || !passMatch {
			rejectBasicAuth(w, cfg)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func rejectBasicAuth(w http.ResponseWriter, cfg *config.BasicAuthConfig) {
	realm := cfg.Realm
	if realm == "" {
		realm = "Restricted"
	}
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	status := cfg.ResolvedStatusCode()
	body := cfg.ResolvedBody()
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

// decodeBasicAuthHeader is a helper used in tests to build a valid
// Authorization header value from a username and password.
func decodeBasicAuthHeader(header string) (user, pass string, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(header, prefix) {
		return
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(header, prefix))
	if err != nil {
		return
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return
	}
	return parts[0], parts[1], true
}
