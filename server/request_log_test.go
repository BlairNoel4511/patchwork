package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequestLog_AddAndAll(t *testing.T) {
	log := NewRequestLog(10)
	log.Add(RequestLogEntry{Method: "GET", Path: "/foo", StatusCode: 200})
	log.Add(RequestLogEntry{Method: "POST", Path: "/bar", StatusCode: 201})

	entries := log.All()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Path != "/foo" || entries[1].Path != "/bar" {
		t.Errorf("unexpected entry paths: %v", entries)
	}
}

func TestRequestLog_MaxSizeEviction(t *testing.T) {
	log := NewRequestLog(3)
	for i := 0; i < 5; i++ {
		log.Add(RequestLogEntry{StatusCode: i})
	}
	entries := log.All()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", len(entries))
	}
	if entries[0].StatusCode != 2 {
		t.Errorf("expected oldest evicted, first entry status = %d", entries[0].StatusCode)
	}
}

func TestRequestLog_Reset(t *testing.T) {
	log := NewRequestLog(10)
	log.Add(RequestLogEntry{Method: "GET"})
	log.Reset()
	if len(log.All()) != 0 {
		t.Error("expected empty log after reset")
	}
}

func TestRequestLogHandler_GET(t *testing.T) {
	log := NewRequestLog(10)
	log.Add(RequestLogEntry{
		Timestamp:  time.Now().UTC(),
		Method:     "GET",
		Path:       "/test",
		StatusCode: 200,
	})

	req := httptest.NewRequest(http.MethodGet, "/patchwork/requests", nil)
	w := httptest.NewRecorder()
	RequestLogHandler(log).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var entries []RequestLogEntry
	if err := json.NewDecoder(w.Body).Decode(&entries); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(entries) != 1 || entries[0].Path != "/test" {
		t.Errorf("unexpected entries: %v", entries)
	}
}

func TestRequestLogHandler_DELETE(t *testing.T) {
	log := NewRequestLog(10)
	log.Add(RequestLogEntry{Method: "GET"})

	req := httptest.NewRequest(http.MethodDelete, "/patchwork/requests", nil)
	w := httptest.NewRecorder()
	RequestLogHandler(log).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if len(log.All()) != 0 {
		t.Error("expected log to be cleared after DELETE")
	}
}

func TestRequestLogHandler_MethodNotAllowed(t *testing.T) {
	log := NewRequestLog(10)
	req := httptest.NewRequest(http.MethodPut, "/patchwork/requests", nil)
	w := httptest.NewRecorder()
	RequestLogHandler(log).ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
