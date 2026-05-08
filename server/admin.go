package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/user/patchwork/config"
)

// AdminRouter registers the built-in admin API onto mux.
// It exposes:
//
//	GET  {prefix}/health        – liveness probe
//	GET  {prefix}/routes        – list registered route paths
//	DELETE {prefix}/requests    – clear the request log
//	POST {prefix}/scenarios/reset – reset all scenario state
func AdminRouter(mux *http.ServeMux, cfg *config.Config, log *RequestLog, store *ScenarioStore) {
	ac := cfg.Admin
	if !ac.IsEnabled() {
		return
	}
	prefix := strings.TrimRight(ac.AdminPrefix(), "/")

	mux.HandleFunc(prefix+"/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc(prefix+"/routes", func(w http.ResponseWriter, r *http.Request) {
		paths := make([]string, 0, len(cfg.Routes))
		for _, route := range cfg.Routes {
			paths = append(paths, route.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"routes": paths})
	})

	mux.HandleFunc(prefix+"/requests", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if ac.ReadOnly {
			http.Error(w, "admin API is read-only", http.StatusForbidden)
			return
		}
		log.Reset()
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc(prefix+"/scenarios/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if ac.ReadOnly {
			http.Error(w, "admin API is read-only", http.StatusForbidden)
			return
		}
		store.Reset()
		w.WriteHeader(http.StatusNoContent)
	})
}
