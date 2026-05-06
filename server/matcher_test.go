package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeMatchRequest(header map[string]string, query map[string]string, body string) *http.Request {
	url := "/test"
	if len(query) > 0 {
		parts := []string{}
		for k, v := range query {
			parts = append(parts, k+"="+v)
		}
		url += "?" + strings.Join(parts, "&")
	}
	r := httptest.NewRequest(http.MethodGet, url, nil)
	for k, v := range header {
		r.Header.Set(k, v)
	}
	if body != "" {
		r = r.WithContext(context.WithValue(r.Context(), bodyKey{}, body))
	}
	return r
}

func TestMatchesRequest_EmptyCondition(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if !matchesRequest(r, MatchCondition{}) {
		t.Error("empty condition should always match")
	}
}

func TestMatchesRequest_HeaderMatch(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Role", "admin")
	cond := MatchCondition{Header: map[string]string{"X-Role": "admin"}}
	if !matchesRequest(r, cond) {
		t.Error("expected header match")
	}
}

func TestMatchesRequest_HeaderMismatch(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Role", "user")
	cond := MatchCondition{Header: map[string]string{"X-Role": "admin"}}
	if matchesRequest(r, cond) {
		t.Error("expected header mismatch")
	}
}

func TestMatchesRequest_QueryMatch(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?env=prod", nil)
	cond := MatchCondition{Query: map[string]string{"env": "prod"}}
	if !matchesRequest(r, cond) {
		t.Error("expected query match")
	}
}

func TestMatchesRequest_BodyContains(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), bodyKey{}, `{"type":"ping"}`))
	cond := MatchCondition{Body: "ping"}
	if !matchesRequest(r, cond) {
		t.Error("expected body match")
	}
}

func TestMatchesRequest_BodyMissingContext(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	cond := MatchCondition{Body: "ping"}
	if matchesRequest(r, cond) {
		t.Error("expected no match when body context missing")
	}
}
