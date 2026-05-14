package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/patrickward/patchwork/config"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// generateID returns a random 16-byte hex string.
func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// RequestIDFromContext retrieves the request ID stored by RequestIDMiddleware.
func RequestIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(requestIDKey).(string)
	return v
}

// RequestIDMiddleware attaches a unique request ID to every request and echoes
// it back in the configured response header.
func RequestIDMiddleware(cfg *config.RequestIDConfig, next http.Handler) http.Handler {
	if !cfg.IsEnabled() {
		return next
	}

	header := cfg.ResolvedHeader()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(header)
		if id == "" || cfg.ForceNew {
			id = generateID()
		}

		// Propagate on the outgoing response.
		w.Header().Set(header, id)

		// Store in context so downstream handlers can read it.
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
