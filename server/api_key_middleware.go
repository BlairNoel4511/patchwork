package server

import (
	"net/http"
	"strings"

	"github.com/patrickward/patchwork/config"
)

// APIKeyMiddleware validates an API key from a header or query parameter.
// If the route has no API key config, the request passes through unchanged.
func APIKeyMiddleware(cfg *config.APIKeyConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg == nil || !cfg.IsEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		key := extractAPIKey(r, cfg)
		if key == "" {
			rejectAPIKey(w, cfg, "missing API key")
			return
		}

		for _, valid := range cfg.Keys {
			if key == valid {
				next.ServeHTTP(w, r)
				return
			}
		}

		rejectAPIKey(w, cfg, "invalid API key")
	})
}

func extractAPIKey(r *http.Request, cfg *config.APIKeyConfig) string {
	if cfg.Header != "" {
		if v := r.Header.Get(cfg.Header); v != "" {
			return strings.TrimSpace(v)
		}
	}
	if cfg.QueryParam != "" {
		if v := r.URL.Query().Get(cfg.QueryParam); v != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func rejectAPIKey(w http.ResponseWriter, cfg *config.APIKeyConfig, msg string) {
	status := cfg.ResolvedStatus()
	body := cfg.ResolvedBody(msg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}
