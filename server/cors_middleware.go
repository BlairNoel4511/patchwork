package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/user/patchwork/config"
)

// RouteCORSMiddleware applies per-route CORS headers defined in the route's
// CORSConfig, overriding or supplementing the global CORSMiddleware.
func RouteCORSMiddleware(cors *config.CORSConfig, next http.Handler) http.Handler {
	if !cors.IsEnabled() {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigin := resolveOrigin(cors, origin)

		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}
		w.Header().Set("Access-Control-Allow-Methods", cors.AllowedMethodsValue())
		w.Header().Set("Access-Control-Allow-Headers", cors.AllowedHeadersValue())

		if len(cors.ExposedHeaders) > 0 {
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(cors.ExposedHeaders, ", "))
		}
		if cors.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if cors.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", cors.MaxAge))
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// resolveOrigin checks whether the request origin is permitted and returns the
// value that should be set in Access-Control-Allow-Origin.
func resolveOrigin(cors *config.CORSConfig, requestOrigin string) string {
	for _, o := range cors.AllowedOrigins {
		if o == "*" {
			return "*"
		}
		if strings.EqualFold(o, requestOrigin) {
			return requestOrigin
		}
	}
	return ""
}
