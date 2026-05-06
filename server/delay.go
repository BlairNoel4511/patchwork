package server

import (
	"net/http"
	"time"
)

// DelayMiddleware adds an artificial delay to responses when a route
// specifies a delay_ms value greater than zero.
func DelayMiddleware(delayMs int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if delayMs > 0 {
				time.Sleep(time.Duration(delayMs) * time.Millisecond)
			}
			next.ServeHTTP(w, r)
		})
	}
}
