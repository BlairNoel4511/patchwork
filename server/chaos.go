package server

import (
	"math/rand"
	"net/http"

	"github.com/user/patchwork/config"
)

// ChaosMiddleware randomly injects failure responses based on route chaos config.
// If a route defines chaos.error_rate (0.0–1.0), requests will randomly receive
// the configured status code and body at that probability.
func ChaosMiddleware(route config.Route) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			chaos := route.Chaos
			if chaos == nil || chaos.ErrorRate <= 0 {
				next.ServeHTTP(w, r)
				return
			}

			if rand.Float64() < chaos.ErrorRate {
				status := chaos.StatusCode
				if status == 0 {
					status = http.StatusInternalServerError
				}
				body := chaos.Body
				if body == "" {
					body = http.StatusText(status)
				}
				w.Header().Set("Content-Type", "text/plain")
				w.Header().Set("X-Patchwork-Chaos", "true")
				w.WriteHeader(status)
				w.Write([]byte(body))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
