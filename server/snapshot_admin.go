package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// SnapshotAdminHandler returns an http.Handler that exposes snapshot
// management endpoints under the given prefix.
//
//	GET  <prefix>/snapshots       — list recorded snapshot keys
//	DELETE <prefix>/snapshots     — delete all snapshots in dir
func SnapshotAdminHandler(dir string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/snapshots", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listSnapshots(w, dir)
		case http.MethodDelete:
			purgeSnapshots(w, dir)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}

func listSnapshots(w http.ResponseWriter, dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]"))
			return
		}
		http.Error(w, "could not read snapshot dir", http.StatusInternalServerError)
		return
	}

	keys := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			keys = append(keys, strings.TrimSuffix(e.Name(), ".json"))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

func purgeSnapshots(w http.ResponseWriter, dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, "could not read snapshot dir", http.StatusInternalServerError)
		return
	}

	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
