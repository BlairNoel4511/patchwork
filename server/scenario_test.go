package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func freshStore() *scenarioStore {
	return &scenarioStore{states: make(map[string]string)}
}

func TestScenarioStore_SetAndGet(t *testing.T) {
	s := freshStore()
	s.Set("route-a", "error")
	if got := s.Get("route-a"); got != "error" {
		t.Fatalf("expected 'error', got %q", got)
	}
}

func TestScenarioStore_GetUnset(t *testing.T) {
	s := freshStore()
	if got := s.Get("missing"); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestScenarioStore_Reset(t *testing.T) {
	s := freshStore()
	s.Set("k", "v")
	s.Reset()
	if got := s.Get("k"); got != "" {
		t.Fatalf("expected empty after reset, got %q", got)
	}
}

func TestScenarioControlHandler_SetsScenario(t *testing.T) {
	store := freshStore()
	h := ScenarioControlHandler(store)
	req := httptest.NewRequest(http.MethodPost, "/_patchwork/scenario?key=myroute&scenario=slow", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if got := store.Get("myroute"); got != "slow" {
		t.Fatalf("expected 'slow', got %q", got)
	}
}

func TestScenarioControlHandler_MissingParams(t *testing.T) {
	store := freshStore()
	h := ScenarioControlHandler(store)
	req := httptest.NewRequest(http.MethodPost, "/_patchwork/scenario?key=myroute", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestScenarioControlHandler_WrongMethod(t *testing.T) {
	store := freshStore()
	h := ScenarioControlHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/_patchwork/scenario?key=k&scenario=s", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestSelectScenarioResponse_MatchesActive(t *testing.T) {
	store := freshStore()
	store.Set("/api/items", "error")
	responses := []Response{
		{Status: 200, Body: "ok", Scenario: ""},
		{Status: 500, Body: "fail", Scenario: "error"},
	}
	got := selectScenarioResponse(store, "/api/items", responses)
	if got == nil || got.Status != 500 {
		t.Fatalf("expected status 500, got %v", got)
	}
}

func TestSelectScenarioResponse_FallbackWhenNoMatch(t *testing.T) {
	store := freshStore()
	responses := []Response{
		{Status: 200, Body: "ok", Scenario: ""},
		{Status: 500, Body: "fail", Scenario: "error"},
	}
	got := selectScenarioResponse(store, "/api/items", responses)
	if got == nil || got.Status != 200 {
		t.Fatalf("expected fallback status 200, got %v", got)
	}
}

func TestSelectScenarioResponse_EmptyList(t *testing.T) {
	store := freshStore()
	got := selectScenarioResponse(store, "/api/items", []Response{})
	if got != nil {
		t.Fatalf("expected nil for empty list, got %v", got)
	}
}
