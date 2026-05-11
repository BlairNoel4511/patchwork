package server

import (
	"net/http"

	"github.com/zachlatta/patchwork/config"
)

// HeaderMiddleware injects static request/response headers defined on a route.
// Request headers are added to the incoming request before passing downstream;
// response headers are written to the response after the handler returns.
func HeaderMiddleware(route config.Route, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Inject extra headers into the outbound request clone so downstream
		// handlers (e.g. proxy) can see them.
		if len(route.RequestHeaders) > 0 {
			r = r.Clone(r.Context())
			for k, v := range route.RequestHeaders {
				r.Header.Set(k, v)
			}
		}

		// Wrap the ResponseWriter so we can append response headers after the
		// handler writes its status line (but before the body is flushed).
		if len(route.ResponseHeaders) > 0 {
			w = &headerInjector{ResponseWriter: w, headers: route.ResponseHeaders}
		}

		next.ServeHTTP(w, r)
	})
}

// headerInjector wraps http.ResponseWriter and injects additional headers on
// the first call to WriteHeader or Write.
type headerInjector struct {
	http.ResponseWriter
	headers map[string]string
	injected bool
}

func (h *headerInjector) inject() {
	if h.injected {
		return
	}
	h.injected = true
	for k, v := range h.headers {
		h.ResponseWriter.Header().Set(k, v)
	}
}

func (h *headerInjector) WriteHeader(code int) {
	h.inject()
	h.ResponseWriter.WriteHeader(code)
}

func (h *headerInjector) Write(b []byte) (int, error) {
	h.inject()
	return h.ResponseWriter.Write(b)
}
