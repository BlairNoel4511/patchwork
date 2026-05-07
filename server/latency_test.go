package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/patrickward/patchwork/config"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestLatencyMiddleware_NilProfile(t *testing.T) {
	h := LatencyMiddleware(nil)(okHandler())
	rr := httptest.NewRecorder()
	start := time.Now()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if time.Since(start) > 50*time.Millisecond {
		t.Error("expected no significant delay with nil profile")
	}
}

func TestLatencyMiddleware_FixedDelay(t *testing.T) {
	profile := &config.LatencyProfile{Distribution: "fixed", FixedMs: 50}
	h := LatencyMiddleware(profile)(okHandler())
	rr := httptest.NewRecorder()
	start := time.Now()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	elapsed := time.Since(start)
	if elapsed < 45*time.Millisecond {
		t.Errorf("expected ~50ms delay, got %v", elapsed)
	}
}

func TestLatencyMiddleware_UniformDelay(t *testing.T) {
	profile := &config.LatencyProfile{Distribution: "uniform", MinMs: 20, MaxMs: 40}
	h := LatencyMiddleware(profile)(okHandler())
	for i := 0; i < 10; i++ {
		rr := httptest.NewRecorder()
		start := time.Now()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
		elapsed := time.Since(start)
		if elapsed < 15*time.Millisecond || elapsed > 100*time.Millisecond {
			t.Errorf("uniform delay out of expected range: %v", elapsed)
		}
	}
}

func TestLatencyMiddleware_NormalDelay(t *testing.T) {
	profile := &config.LatencyProfile{Distribution: "normal", MeanMs: 30, StdDevMs: 5}
	h := LatencyMiddleware(profile)(okHandler())
	rr := httptest.NewRecorder()
	start := time.Now()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	elapsed := time.Since(start)
	// Allow wide margin for normal distribution
	if elapsed > 200*time.Millisecond {
		t.Errorf("normal delay unexpectedly large: %v", elapsed)
	}
}

func TestSampleLatency_ZeroFixed(t *testing.T) {
	p := &config.LatencyProfile{Distribution: "fixed", FixedMs: 0}
	if d := sampleLatency(p); d != 0 {
		t.Errorf("expected 0, got %v", d)
	}
}

func TestSampleLatency_UniformZeroSpan(t *testing.T) {
	p := &config.LatencyProfile{Distribution: "uniform", MinMs: 10, MaxMs: 10}
	if d := sampleLatency(p); d != 10*time.Millisecond {
		t.Errorf("expected 10ms, got %v", d)
	}
}

func TestSampleLatency_NormalNonNegative(t *testing.T) {
	p := &config.LatencyProfile{Distribution: "normal", MeanMs: 0, StdDevMs: 1}
	for i := 0; i < 100; i++ {
		if d := sampleLatency(p); d < 0 {
			t.Errorf("expected non-negative duration, got %v", d)
		}
	}
}
