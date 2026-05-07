package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/patrickward/patchwork/config"
)

// TestLatencyMiddleware_ChainedWithHandler verifies latency middleware
// passes the response through correctly when chained with a real handler.
func TestLatencyMiddleware_ChainedWithHandler(t *testing.T) {
	profile := &config.LatencyProfile{Distribution: "fixed", FixedMs: 10}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "yes")
		w.WriteHeader(http.StatusCreated)
	})

	h := LatencyMiddleware(profile)(inner)
	rr := httptest.NewRecorder()
	start := time.Now()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/data", nil))
	elapsed := time.Since(start)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
	if rr.Header().Get("X-Test") != "yes" {
		t.Error("expected X-Test header to be forwarded")
	}
	if elapsed < 5*time.Millisecond {
		t.Errorf("expected at least 10ms delay, got %v", elapsed)
	}
}
