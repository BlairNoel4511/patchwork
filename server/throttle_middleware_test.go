package server

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/patchwork/config"
)

func makeThrottleRoute(max, queue int) *config.ThrottleConfig {
	return &config.ThrottleConfig{
		MaxConcurrent: max,
		QueueSize:     queue,
		QueueTimeout:  "200ms",
	}
}

func okThrottleHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestThrottleMiddleware_NoConfig_PassesThrough(t *testing.T) {
	h := ThrottleMiddleware(nil, okThrottleHandler())
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestThrottleMiddleware_AllowsWithinLimit(t *testing.T) {
	cfg := makeThrottleRoute(5, 0)
	h := ThrottleMiddleware(cfg, okThrottleHandler())
	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, rr.Code)
		}
	}
}

func TestThrottleMiddleware_RejectsWhenFull(t *testing.T) {
	blocking := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocking
		w.WriteHeader(http.StatusOK)
	})
	cfg := &config.ThrottleConfig{MaxConcurrent: 1, StatusCode: 503, Body: "busy"}
	h := ThrottleMiddleware(cfg, handler)

	var started sync.WaitGroup
	started.Add(1)
	go func() {
		started.Done()
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}()
	started.Wait()
	time.Sleep(20 * time.Millisecond)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	close(blocking)

	if rr.Code != 503 {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
}

func TestThrottleMiddleware_QueuedRequestEventuallyServed(t *testing.T) {
	blocking := make(chan struct{})
	var served atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocking
		served.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	cfg := &config.ThrottleConfig{MaxConcurrent: 1, QueueSize: 2, QueueTimeout: "300ms"}
	h := ThrottleMiddleware(cfg, handler)

	var wg sync.WaitGroup
	results := make([]int, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
			results[idx] = rr.Code
		}(i)
	}
	time.Sleep(30 * time.Millisecond)
	close(blocking)
	wg.Wait()

	for i, code := range results {
		if code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, code)
		}
	}
}

func TestThrottleMiddleware_CustomStatusAndBody(t *testing.T) {
	cfg := &config.ThrottleConfig{MaxConcurrent: 0, StatusCode: 429, Body: "rate limited"}
	// MaxConcurrent=0 means disabled, so it passes through — test via direct reject path.
	cfg2 := &config.ThrottleConfig{MaxConcurrent: 1, StatusCode: 429, Body: "rate limited"}
	blocking := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-blocking })
	h := ThrottleMiddleware(cfg2, handler)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}()
	time.Sleep(20 * time.Millisecond)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	close(blocking)
	wg.Wait()

	if rr.Code != 429 {
		t.Fatalf("expected 429, got %d", rr.Code)
	}
	if body := rr.Body.String(); body != "rate limited" {
		t.Fatalf("unexpected body: %q", body)
	}
	_ = cfg
}
