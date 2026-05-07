package server

import (
	"net/http"
	"sync"
	"time"

	"github.com/user/patchwork/config"
)

// tokenBucket implements a simple per-route token bucket rate limiter.
type tokenBucket struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per nanosecond
	lastTick time.Time
}

func newTokenBucket(rps float64, burst int) *tokenBucket {
	return &tokenBucket{
		tokens:   float64(burst),
		max:      float64(burst),
		rate:     rps / float64(time.Second),
		lastTick: time.Now(),
	}
}

func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(tb.lastTick)
	tb.lastTick = now
	tb.tokens += float64(elapsed) * tb.rate
	if tb.tokens > tb.max {
		tb.tokens = tb.max
	}
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// rateLimitStore holds per-route token buckets keyed by route path+method.
type rateLimitStore struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
}

var globalRateLimitStore = &rateLimitStore{
	buckets: make(map[string]*tokenBucket),
}

func (s *rateLimitStore) get(key string, rl config.RateLimit) *tokenBucket {
	s.mu.Lock()
	defer s.mu.Unlock()
	if b, ok := s.buckets[key]; ok {
		return b
	}
	burst := rl.Burst
	if burst <= 0 {
		burst = 1
	}
	b := newTokenBucket(rl.RequestsPerSecond, burst)
	s.buckets[key] = b
	return b
}

// RateLimitMiddleware enforces rate limiting for a route if configured.
func RateLimitMiddleware(route config.Route) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rl := route.RateLimit
			if rl == nil || rl.RequestsPerSecond <= 0 {
				next.ServeHTTP(w, r)
				return
			}
			key := route.Method + ":" + route.Path
			bucket := globalRateLimitStore.get(key, *rl)
			if !bucket.allow() {
				status := rl.StatusCode
				if status == 0 {
					status = http.StatusTooManyRequests
				}
				body := rl.Body
				if body == "" {
					body = "rate limit exceeded"
				}
				http.Error(w, body, status)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
