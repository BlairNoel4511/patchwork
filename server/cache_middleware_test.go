package server

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jbclayton/patchwork/config"
)

func resetCache() {
	cacheMu.Lock()
	cacheStore = map[string]cacheEntry{}
	cacheMu.Unlock()
}

func TestCacheMiddleware_Disabled(t *testing.T) {
	resetCache()
	var calls int32
	h := CacheMiddleware(&config.CacheConfig{Enabled: false}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	if calls != 3 {
		t.Fatalf("expected 3 handler calls when disabled, got %d", calls)
	}
}

func TestCacheMiddleware_CachesResponse(t *testing.T) {
	resetCache()
	var calls int32
	h := CacheMiddleware(&config.CacheConfig{Enabled: true, TTL: "10s"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`)) //nolint:errcheck
	}))

	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/data", nil))
	if rec1.Header().Get("X-Cache") != "MISS" {
		t.Fatalf("first request should be MISS")
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/data", nil))
	if rec2.Header().Get("X-Cache") != "HIT" {
		t.Fatalf("second request should be HIT")
	}
	if calls != 1 {
		t.Fatalf("handler should only be called once, got %d", calls)
	}
}

func TestCacheMiddleware_VaryByQuery(t *testing.T) {
	resetCache()
	var calls int32
	h := CacheMiddleware(&config.CacheConfig{Enabled: true, TTL: "10s", VaryByQuery: []string{"page"}},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(http.StatusOK)
		}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/items?page=1", nil))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/items?page=2", nil))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/items?page=1", nil))

	if calls != 2 {
		t.Fatalf("expected 2 unique cache keys (page=1, page=2), got %d calls", calls)
	}
}

func TestCacheMiddleware_Expiry(t *testing.T) {
	resetCache()
	var calls int32
	h := CacheMiddleware(&config.CacheConfig{Enabled: true, TTL: "50ms"},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(http.StatusOK)
		}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/exp", nil))
	time.Sleep(80 * time.Millisecond)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/exp", nil))

	if calls != 2 {
		t.Fatalf("expected 2 handler calls after TTL expiry, got %d", calls)
	}
}
