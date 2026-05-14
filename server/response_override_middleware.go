package server

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/patrickward/patchwork/config"
)

// overrideStore holds active response overrides keyed by route path.
type overrideStore struct {
	mu    sync.RWMutex
	store map[string]*config.ResponseOverride
}

var globalOverrideStore = &overrideStore{
	store: make(map[string]*config.ResponseOverride),
}

func (s *overrideStore) set(path string, o *config.ResponseOverride) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[path] = o
}

func (s *overrideStore) get(path string) (*config.ResponseOverride, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.store[path]
	return v, ok
}

func (s *overrideStore) delete(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, path)
}

func (s *overrideStore) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store = make(map[string]*config.ResponseOverride)
}

// ResponseOverrideMiddleware checks whether an active override exists for the
// current route path and, if so, short-circuits the handler chain with the
// overridden status/body/headers.
func ResponseOverrideMiddleware(routePath string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if o, ok := globalOverrideStore.get(routePath); ok && !o.IsEmpty() {
			for k, v := range o.Headers {
				w.Header().Set(k, v)
			}
			status := o.Status
			if status == 0 {
				status = http.StatusOK
			}
			w.WriteHeader(status)
			_, _ = w.Write([]byte(o.Body))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ResponseOverrideAdminHandler provides REST endpoints to manage overrides.
//
//	PUT  /__admin/overrides/{path}  — set an override
//	DELETE /__admin/overrides/{path} — clear an override
//	DELETE /__admin/overrides        — clear all overrides
func ResponseOverrideAdminHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		routePath := r.URL.Query().Get("route")
		switch r.Method {
		case http.MethodPut:
			if routePath == "" {
				http.Error(w, "missing route query param", http.StatusBadRequest)
				return
			}
			var o config.ResponseOverride
			if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
			globalOverrideStore.set(routePath, &o)
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if routePath == "" {
				globalOverrideStore.reset()
			} else {
				globalOverrideStore.delete(routePath)
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
