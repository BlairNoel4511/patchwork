package server

import (
	"fmt"
	"net/http"

	"github.com/user/patchwork/config"
)

// New builds and returns an http.Handler from the provided config.
func New(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()
	store := NewScenarioStore()
	log := NewRequestLog(cfg.Server.MaxLogEntries)

	for _, route := range cfg.Routes {
		r := route // capture
		var handler http.Handler
		if len(r.Conditions) > 0 {
			handler = NewConditionalHandler(r)
		} else {
			handler = NewRouteHandler(r)
		}
		if r.Scenario != "" {
			handler = scenarioHandler(store, r, handler)
		}
		chain := Chain(
			handler,
			BodyCacheMiddleware,
			RateLimitMiddleware(r),
			DelayMiddleware(r),
			ChaosMiddleware(r),
			ProxyMiddleware(r),
			WebhookMiddleware(r),
			RequestLogMiddleware(log),
			LoggingMiddleware,
			CORSMiddleware,
		)
		mux.Handle(fmt.Sprintf("%s %s", r.Method, r.Path), chain)
	}

	mux.Handle("GET /__patchwork/scenarios", ScenarioControlHandler(store))
	mux.Handle("POST /__patchwork/scenarios", ScenarioControlHandler(store))
	mux.Handle("DELETE /__patchwork/scenarios", ScenarioControlHandler(store))
	mux.Handle("GET /__patchwork/requests", RequestLogHandler(log))
	mux.Handle("DELETE /__patchwork/requests", RequestLogHandler(log))

	return mux
}
