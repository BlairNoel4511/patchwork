package server

import (
	"net/http"

	"github.com/patrickward/patchwork/config"
)

// applyLatency wraps the given handler with LatencyMiddleware when a
// LatencyProfile is defined on the route. Returns the handler unchanged
// when no profile is configured.
func applyLatency(h http.Handler, route config.Route) http.Handler {
	if route.LatencyProfile == nil {
		return h
	}
	return LatencyMiddleware(route.LatencyProfile)(h)
}
