package server

import (
	"net/http"
	"sync"
)

// scenarioStore holds named scenario states per route group.
type scenarioStore struct {
	mu     sync.RWMutex
	states map[string]string
}

var globalScenarios = &scenarioStore{
	states: make(map[string]string),
}

// Set updates the active scenario for a given key.
func (s *scenarioStore) Set(key, scenario string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[key] = scenario
}

// Get returns the active scenario for a given key, or empty string if unset.
func (s *scenarioStore) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.states[key]
}

// Reset clears all scenario state.
func (s *scenarioStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states = make(map[string]string)
}

// ScenarioControlHandler returns an HTTP handler that allows clients to set
// the active scenario for a route group via POST /_patchwork/scenario.
//
// Expected query params:
//   - key:      the scenario group identifier (e.g. route path)
//   - scenario: the scenario name to activate
func ScenarioControlHandler(store *scenarioStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		key := r.URL.Query().Get("key")
		scenario := r.URL.Query().Get("scenario")
		if key == "" || scenario == "" {
			http.Error(w, "missing key or scenario query param", http.StatusBadRequest)
			return
		}
		store.Set(key, scenario)
		w.WriteHeader(http.StatusNoContent)
	}
}

// selectScenarioResponse picks the response whose scenario tag matches the
// active scenario for the route key. Falls back to the first untagged response
// if no match is found.
func selectScenarioResponse(store *scenarioStore, routeKey string, responses []Response) *Response {
	active := store.Get(routeKey)
	var fallback *Response
	for i := range responses {
		if responses[i].Scenario == active && active != "" {
			return &responses[i]
		}
		if fallback == nil && responses[i].Scenario == "" {
			fallback = &responses[i]
		}
	}
	if fallback != nil {
		return fallback
	}
	if len(responses) > 0 {
		return &responses[0]
	}
	return nil
}
