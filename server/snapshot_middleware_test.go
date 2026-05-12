package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/patrickward/patchwork/config"
)

func boolPtr(b bool) *bool { return &b }

func makeSnapshotConfig(dir string, enabled bool, ttl string) *config.SnapshotConfig {
	return &config.SnapshotConfig{
		Enabled: boolPtr(enabled),
		Dir:     dir,
		TTL:     ttl,
	}
}

func TestSnapshotMiddleware_Disabled(t *testing.T) {
	called := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusOK)
	})
	h := SnapshotMiddleware(makeSnapshotConfig("", false, ""), next)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if called != 1 {
		t.Fatalf("expected next to be called once, got %d", called)
	}
}

func TestSnapshotMiddleware_RecordsAndReplays(t *testing.T) {
	dir := t.TempDir()
	called := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})
	h := SnapshotMiddleware(makeSnapshotConfig(dir, true, ""), next)

	// First request — should call next and save snapshot.
	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/items", nil))
	if called != 1 {
		t.Fatalf("expected 1 upstream call, got %d", called)
	}

	// Second request — should replay from snapshot.
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/items", nil))
	if called != 1 {
		t.Fatalf("expected no additional upstream call, got %d", called)
	}
	if rec2.Header().Get("X-Patchwork-Snapshot") != "hit" {
		t.Fatal("expected snapshot hit header")
	}
	if rec2.Body.String() != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", rec2.Body.String())
	}
}

func TestSnapshotMiddleware_TTLExpiry(t *testing.T) {
	dir := t.TempDir()
	called := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	})
	h := SnapshotMiddleware(makeSnapshotConfig(dir, true, "1ms"), next)

	req := httptest.NewRequest(http.MethodGet, "/ttl", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	time.Sleep(5 * time.Millisecond)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if called != 2 {
		t.Fatalf("expected 2 upstream calls after TTL expiry, got %d", called)
	}
}

func TestSnapshotKey_Deterministic(t *testing.T) {
	r1 := httptest.NewRequest(http.MethodGet, "/foo?bar=1", nil)
	r2 := httptest.NewRequest(http.MethodGet, "/foo?bar=1", nil)
	if snapshotKey(r1) != snapshotKey(r2) {
		t.Fatal("keys should be equal for identical requests")
	}
}

func TestSaveAndLoadSnapshot(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	entry := snapshotEntry{
		Status:  201,
		Headers: map[string]string{"Content-Type": "text/plain"},
		Body:    "created",
		SavedAt: time.Now(),
	}
	data, _ := json.MarshalIndent(entry, "", "  ")
	os.WriteFile(path, data, 0o644)

	loaded, ok := loadSnapshot(path, 0)
	if !ok {
		t.Fatal("expected snapshot to load")
	}
	if loaded.Status != 201 || loaded.Body != "created" {
		t.Fatalf("unexpected snapshot content: %+v", loaded)
	}
}
