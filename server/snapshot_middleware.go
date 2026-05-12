package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/patrickward/patchwork/config"
)

type snapshotEntry struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	SavedAt time.Time         `json:"saved_at"`
}

// SnapshotMiddleware records the first response for a route to disk and
// replays it on subsequent requests until the TTL expires.
func SnapshotMiddleware(cfg *config.SnapshotConfig, next http.Handler) http.Handler {
	if !cfg.IsEnabled() {
		return next
	}
	ttl, err := cfg.ParseTTL()
	if err != nil {
		return next
	}
	dir := cfg.SnapshotDir()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := snapshotKey(r)
		path := filepath.Join(dir, key+".json")

		if entry, ok := loadSnapshot(path, ttl); ok {
			for k, v := range entry.Headers {
				w.Header().Set(k, v)
			}
			w.Header().Set("X-Patchwork-Snapshot", "hit")
			w.WriteHeader(entry.Status)
			fmt.Fprint(w, entry.Body)
			return
		}

		rec := newResponseWriter(w)
		next.ServeHTTP(rec, r)

		entry := snapshotEntry{
			Status:  rec.status,
			Headers: flattenHeaders(rec.Header()),
			Body:    rec.body.String(),
			SavedAt: time.Now(),
		}
		_ = saveSnapshot(path, entry)
	})
}

func snapshotKey(r *http.Request) string {
	raw := r.Method + "|" + r.URL.Path + "|" + r.URL.RawQuery
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum[:8])
}

func loadSnapshot(path string, ttl time.Duration) (*snapshotEntry, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var entry snapshotEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}
	if ttl > 0 && time.Since(entry.SavedAt) > ttl {
		_ = os.Remove(path)
		return nil, false
	}
	return &entry, true
}

func saveSnapshot(path string, entry snapshotEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func flattenHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for k, vs := range h {
		if len(vs) > 0 {
			out[k] = vs[0]
		}
	}
	return out
}

// ensure responseWriter exposes body for snapshot capture
type capturingWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (c *capturingWriter) WriteHeader(code int) {
	c.status = code
	c.ResponseWriter.WriteHeader(code)
}

func (c *capturingWriter) Write(b []byte) (int, error) {
	c.body.Write(b)
	return c.ResponseWriter.Write(b)
}
