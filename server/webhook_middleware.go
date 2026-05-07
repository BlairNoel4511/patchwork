package server

import (
	"net/http"

	"github.com/user/patchwork/config"
)

// WebhookMiddleware fires outbound webhooks after the inner handler writes
// its response. It wraps the dispatcher so routes stay unaware of dispatch
// mechanics.
func WebhookMiddleware(route config.Route, dispatcher *WebhookDispatcher) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)

			hooks := routeWebhooks(route)
			if len(hooks) == 0 {
				return
			}
			dispatcher.Dispatch(r.Context(), hooks)
		})
	}
}

// routeWebhooks collects webhooks from the top-level route definition.
// Response-level webhooks are intentionally not supported here; they would
// be resolved by the handler after response selection.
func routeWebhooks(route config.Route) []config.Webhook {
	return route.Webhooks
}
