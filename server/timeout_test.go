package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/patchwork/config"
)

func slowHandler(d time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(d):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		case <-r.Context().Done():
			// context cancelled — do nothing
		}
	})
}

func TestTimeoutMiddleware_NoConfig_PassesThrough(t *testing.T) {
	cfg := &config.TimeoutConfig{} // empty duration => no timeout
	h := TimeoutMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("hello"))
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "hello" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestTimeoutMiddleware_HandlerCompletesInTime(t *testing.T) {
	cfg := &config.TimeoutConfig{Duration: "500ms"}
	h := TimeoutMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("fast"))
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestTimeoutMiddleware_ExceedsDeadline(t *testing.T) {
	cfg := &config.TimeoutConfig{Duration: "50ms"}
	h := TimeoutMiddleware(cfg, slowHandler(300*time.Millisecond))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != 504 {
		t.Fatalf("expected 504, got %d", rec.Code)
	}
	if rec.Body.String() != "gateway timeout" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestTimeoutMiddleware_CustomStatusAndBody(t *testing.T) {
	cfg := &config.TimeoutConfig{
		Duration:   "30ms",
		StatusCode: 408,
		Body:       "too slow",
	}
	h := TimeoutMiddleware(cfg, slowHandler(500*time.Millisecond))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != 408 {
		t.Fatalf("expected 408, got %d", rec.Code)
	}
	if rec.Body.String() != "too slow" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}
