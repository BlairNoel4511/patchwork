package server

import (
	"net/http"

	"github.com/user/patchwork/config"
)

// applyThrottle wraps next with ThrottleMiddleware when the route has a
// throttle configuration, otherwise it returns next unchanged.
func applyThrottle(route config.Route, next http.Handler) http.Handler {
	if route.Throttle == nil || !route.Throttle.IsEnabled() {
		return next
	}
	return ThrottleMiddleware(route.Throttle, next)
}
