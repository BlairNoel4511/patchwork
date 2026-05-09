package server

import (
	"bytes"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/patrickward/patchwork/config"
)

// RetryMiddleware retries the next handler when it returns a status code that
// matches the route's retry configuration. It captures the response via a
// recorder and only flushes to the real ResponseWriter on the final attempt.
func RetryMiddleware(cfg *config.RetryConfig, next http.Handler) http.Handler {
	if !cfg.IsEnabled() {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			rec    *retryRecorder
			delay = time.Duration(cfg.BackoffMS) * time.Millisecond
		)
		for attempt := 0; attempt < cfg.Attempts; attempt++ {
			if attempt > 0 {
				log.Printf("[retry] attempt %d/%d after %v", attempt+1, cfg.Attempts, delay)
				time.Sleep(delay)
				delay = nextDelay(delay, cfg)
			}
			rec = &retryRecorder{header: w.Header().Clone(), code: http.StatusOK}
			next.ServeHTTP(rec, r)
			if !cfg.ShouldRetry(rec.code) {
				break
			}
		}
		// Flush the final recorded response to the real writer.
		for k, vals := range rec.header {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(rec.code)
		_, _ = w.Write(rec.body.Bytes())
	})
}

func nextDelay(current time.Duration, cfg *config.RetryConfig) time.Duration {
	mul := cfg.Multiplier
	if mul <= 0 {
		mul = 1
	}
	next := time.Duration(math.Round(float64(current) * mul))
	if cfg.MaxBackoffMS > 0 {
		max := time.Duration(cfg.MaxBackoffMS) * time.Millisecond
		if next > max {
			return max
		}
	}
	return next
}

// retryRecorder is a minimal ResponseWriter that buffers the response.
type retryRecorder struct {
	header http.Header
	code   int
	body   bytes.Buffer
}

func (r *retryRecorder) Header() http.Header        { return r.header }
func (r *retryRecorder) WriteHeader(code int)        { r.code = code }
func (r *retryRecorder) Write(b []byte) (int, error) { return r.body.Write(b) }
