package server

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jbclayton/patchwork/config"
)

type cacheEntry struct {
	status  int
	headers http.Header
	body    []byte
	expiresAt time.Time
}

type responseCapture struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
}

func (rc *responseCapture) WriteHeader(code int) { rc.status = code; rc.ResponseWriter.WriteHeader(code) }
func (rc *responseCapture) Write(b []byte) (int, error) {
	rc.buf.Write(b)
	return rc.ResponseWriter.Write(b)
}

var (
	cacheMu    sync.RWMutex
	cacheStore = map[string]cacheEntry{}
)

// CacheMiddleware caches route responses in-memory according to CacheConfig.
func CacheMiddleware(cfg *config.CacheConfig, next http.Handler) http.Handler {
	if !cfg.IsEnabled() {
		return next
	}
	ttl, err := cfg.ParseTTL()
	if err != nil {
		ttl = 60 * time.Second
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := cacheKey(r, cfg)

		cacheMu.RLock()
		entry, found := cacheStore[key]
		cacheMu.RUnlock()

		if found && time.Now().Before(entry.expiresAt) {
			for k, vals := range entry.headers {
				for _, v := range vals {
					w.Header().Add(k, v)
				}
			}
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(entry.status)
			w.Write(entry.body) //nolint:errcheck
			return
		}

		rc := &responseCapture{ResponseWriter: w, status: http.StatusOK}
		w.Header().Set("X-Cache", "MISS")
		next.ServeHTTP(rc, r)

		cacheMu.Lock()
		cacheStore[key] = cacheEntry{
			status:    rc.status,
			headers:   w.Header().Clone(),
			body:      rc.buf.Bytes(),
			expiresAt: time.Now().Add(ttl),
		}
		cacheMu.Unlock()
	})
}

func cacheKey(r *http.Request, cfg *config.CacheConfig) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s %s", r.Method, r.URL.Path)

	params := r.URL.Query()
	keys := make([]string, 0, len(cfg.VaryByQuery))
	for _, k := range cfg.VaryByQuery {
		keys = append(keys, k+"="+strings.Join(params[k], ","))
	}
	sort.Strings(keys)
	fmt.Fprintf(h, "?%s", strings.Join(keys, "&"))

	hdrParts := make([]string, 0, len(cfg.VaryByHeader))
	for _, k := range cfg.VaryByHeader {
		hdrParts = append(hdrParts, k+":"+r.Header.Get(k))
	}
	sort.Strings(hdrParts)
	fmt.Fprintf(h, " H:%s", strings.Join(hdrParts, ","))

	return fmt.Sprintf("%x", h.Sum(nil))
}
