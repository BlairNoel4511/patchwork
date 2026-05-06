package server

import (
	"context"
	"io"
	"net/http"
)

// BodyCacheMiddleware reads the request body once and stores it in the
// request context so that matchers and templates can access it without
// consuming the reader.
func BodyCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil || r.ContentLength == 0 {
			next.ServeHTTP(w, r)
			return
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusInternalServerError)
			return
		}
		_ = r.Body.Close()

		ctx := context.WithValue(r.Context(), bodyKey{}, string(data))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
