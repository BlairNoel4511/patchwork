package server

import (
	"net/http"
	"sync"
	"time"

	"github.com/damiensedgwick/patchwork/config"
)

type circuitState int

const (
	stateClosed   circuitState = iota
	stateOpen     circuitState = iota
	stateHalfOpen circuitState = iota
)

type circuitBreaker struct {
	mu           sync.Mutex
	state        circuitState
	failures     int
	threshold    int
	openUntil    time.Time
	openDuration time.Duration
}

func newCircuitBreaker(cfg config.CircuitBreakerConfig) *circuitBreaker {
	threshold := cfg.Threshold
	if threshold <= 0 {
		threshold = 5
	}
	openDuration := cfg.OpenDuration
	if openDuration <= 0 {
		openDuration = 5000
	}
	return &circuitBreaker{
		state:        stateClosed,
		threshold:    threshold,
		openDuration: time.Duration(openDuration) * time.Millisecond,
	}
}

func (cb *circuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case stateOpen:
		if time.Now().After(cb.openUntil) {
			cb.state = stateHalfOpen
			return true
		}
		return false
	default:
		return true
	}
}

func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = stateClosed
}

func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.failures >= cb.threshold {
		cb.state = stateOpen
		cb.openUntil = time.Now().Add(cb.openDuration)
	}
}

// CircuitBreakerMiddleware wraps a handler and trips the circuit when the
// downstream handler returns 5xx responses too frequently.
func CircuitBreakerMiddleware(cfg config.CircuitBreakerConfig, cb *circuitBreaker) func(http.Handler) http.Handler {
	statusCode := cfg.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusServiceUnavailable
	}
	body := cfg.Body
	if body == "" {
		body = "circuit breaker open"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cb.allow() {
				w.WriteHeader(statusCode)
				_, _ = w.Write([]byte(body))
				return
			}
			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)
			if rw.status >= 500 {
				cb.recordFailure()
			} else {
				cb.recordSuccess()
			}
		})
	}
}
