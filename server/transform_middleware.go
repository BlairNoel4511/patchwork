package server

import (
	"net/http"
	"strings"

	"github.com/patrickward/patchwork/config"
)

// TransformMiddleware applies request and response transformations defined
// in the route's TransformConfig.
func TransformMiddleware(cfg *config.TransformConfig, next http.Handler) http.Handler {
	if cfg == nil || !cfg.IsEnabled() {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg.Request != nil {
			applyRequestTransform(r, cfg.Request)
		}

		if cfg.Response != nil {
			rw := newResponseWriter(w)
			applyResponseTransform(rw, cfg.Response)
			next.ServeHTTP(rw, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func applyRequestTransform(r *http.Request, rt *config.RequestTransform) {
	if rt.StripPathPrefix != "" && strings.HasPrefix(r.URL.Path, rt.StripPathPrefix) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, rt.StripPathPrefix)
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
	}
	if rt.AddPathPrefix != "" {
		r.URL.Path = rt.AddPathPrefix + r.URL.Path
	}
	for k, v := range rt.SetHeaders {
		r.Header.Set(k, v)
	}
	for _, k := range rt.RemoveHeaders {
		r.Header.Del(k)
	}
}

func applyResponseTransform(rw *responseWriter, rt *config.ResponseTransform) {
	for k, v := range rt.SetHeaders {
		rw.Header().Set(k, v)
	}
	for _, k := range rt.RemoveHeaders {
		rw.Header().Del(k)
	}
	if rt.OverrideStatus != 0 {
		rw.WriteHeader(rt.OverrideStatus)
	}
}
