package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// RequestLogEntry holds details about a single recorded request.
type RequestLogEntry struct {
	Timestamp  time.Time         `json:"timestamp"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Query      string            `json:"query,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
	StatusCode int               `json:"status_code"`
}

// RequestLog is a thread-safe in-memory log of recent requests.
type RequestLog struct {
	mu      sync.RWMutex
	entries []RequestLogEntry
	maxSize int
}

// NewRequestLog creates a RequestLog with the given maximum entry count.
func NewRequestLog(maxSize int) *RequestLog {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &RequestLog{maxSize: maxSize}
}

// Add appends an entry, evicting the oldest if the log is full.
func (l *RequestLog) Add(entry RequestLogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.entries) >= l.maxSize {
		l.entries = l.entries[1:]
	}
	l.entries = append(l.entries, entry)
}

// All returns a copy of all log entries.
func (l *RequestLog) All() []RequestLogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]RequestLogEntry, len(l.entries))
	copy(result, l.entries)
	return result
}

// Reset clears all log entries.
func (l *RequestLog) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = nil
}

// RequestLogHandler returns an HTTP handler that exposes the log as JSON.
// DELETE /patchwork/requests clears the log.
func RequestLogHandler(log *RequestLog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			log.Reset()
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			entries := log.All()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(entries)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// RequestLogMiddleware records each request into the provided RequestLog.
func RequestLogMiddleware(log *RequestLog) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)

			headers := make(map[string]string, len(r.Header))
			for k := range r.Header {
				headers[k] = r.Header.Get(k)
			}

			body, _ := r.Context().Value(bodyKey).(string)

			log.Add(RequestLogEntry{
				Timestamp:  time.Now().UTC(),
				Method:     r.Method,
				Path:       r.URL.Path,
				Query:      r.URL.RawQuery,
				Headers:    headers,
				Body:       body,
				StatusCode: rw.status,
			})
		})
	}
}
