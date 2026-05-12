package server

import (
	"net/http"
	"time"

	"github.com/user/patchwork/config"
)

// ThrottleMiddleware limits the number of concurrent requests handled by next.
// Requests that exceed MaxConcurrent are queued up to QueueSize; if the queue
// is also full they are immediately rejected with the configured status code.
func ThrottleMiddleware(cfg *config.ThrottleConfig, next http.Handler) http.Handler {
	if !cfg.IsEnabled() {
		return next
	}

	sem := make(chan struct{}, cfg.MaxConcurrent)
	var queue chan struct{}
	queueTimeout := 500 * time.Millisecond

	qs := cfg.ResolvedQueueSize()
	if qs > 0 {
		queue = make(chan struct{}, qs)
	}
	if cfg.QueueTimeout != "" {
		if d, err := time.ParseDuration(cfg.QueueTimeout); err == nil {
			queueTimeout = d
		}
	}

	reject := func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(cfg.ResolvedStatusCode())
		_, _ = w.Write([]byte(cfg.ResolvedBody()))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to acquire a slot immediately.
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
			next.ServeHTTP(w, r)
			return
		default:
		}

		// No slot available — try the queue if configured.
		if queue == nil {
			reject(w)
			return
		}

		select {
		case queue <- struct{}{}:
			defer func() { <-queue }()
		default:
			// Queue also full.
			reject(w)
			return
		}

		// Wait for a semaphore slot within the timeout.
		timer := time.NewTimer(queueTimeout)
		defer timer.Stop()
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
			next.ServeHTTP(w, r)
		case <-timer.C:
			reject(w)
		case <-r.Context().Done():
			reject(w)
		}
	})
}
