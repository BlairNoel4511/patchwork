package server

import (
	"net/http"

	"github.com/patrickward/patchwork/config"
)

// applyJWT wraps next with JWTMiddleware when JWT auth is configured on the route.
func applyJWT(route config.Route, next http.Handler) http.Handler {
	if route.JWT == nil || !route.JWT.IsEnabled() {
		return next
	}
	return JWTMiddleware(route.JWT, next)
}
