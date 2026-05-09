package server

import (
	"context"
	"net/http"
	"time"

	"github.com/user/patchwork/config"
)

// TimeoutMiddleware wraps the next handler with a per-request deadline derived
// from the route's TimeoutConfig. If the handler does not complete within the
// deadline, a configurable error response is written and the context is
// cancelled.
func TimeoutMiddleware(cfg *config.TimeoutConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := cfg.Timeout()
		if d <= 0 {
			next.ServeHTTP(w, r)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), d)
		defer cancel()

		done := make(chan struct{})
		prw := &pausableResponseWriter{ResponseWriter: w}

		go func() {
			defer close(done)
			next.ServeHTTP(prw, r.WithContext(ctx))
		}()

		select {
		case <-done:
			prw.flush(w)
		case <-ctx.Done():
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(cfg.ResponseStatus())
			_, _ = w.Write([]byte(cfg.ResponseBody()))
		}
	})
}

// pausableResponseWriter buffers the response so we can discard it on timeout.
type pausableResponseWriter struct {
	http.ResponseWriter
	code int
	header http.Header
	body   []byte
	written bool
}

func (p *pausableResponseWriter) WriteHeader(code int) {
	p.code = code
	p.written = true
}

func (p *pausableResponseWriter) Write(b []byte) (int, error) {
	p.body = append(p.body, b...)
	p.written = true
	return len(b), nil
}

func (p *pausableResponseWriter) Header() http.Header {
	if p.header == nil {
		p.header = make(http.Header)
	}
	return p.header
}

func (p *pausableResponseWriter) flush(w http.ResponseWriter) {
	for k, vs := range p.header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	if p.code != 0 {
		w.WriteHeader(p.code)
	}
	if len(p.body) > 0 {
		_, _ = w.Write(p.body)
	}
}

// applyTimeout wires TimeoutMiddleware only when a config is present.
func applyTimeout(cfg *config.TimeoutConfig, next http.Handler) http.Handler {
	if cfg == nil || cfg.Duration == "" {
		return next
	}
	return TimeoutMiddleware(cfg, next)
}

// ensure applyTimeout is used so the compiler is happy during wiring.
var _ = applyTimeout
var _ = time.Second // keep time import used
