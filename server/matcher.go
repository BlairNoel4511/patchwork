package server

import (
	"net/http"
	"strings"
)

// MatchCondition defines criteria for conditional response matching.
type MatchCondition struct {
	Header map[string]string `yaml:"header,omitempty"`
	Query  map[string]string `yaml:"query,omitempty"`
	Body   string            `yaml:"body,omitempty"`
}

// matchesRequest returns true if the request satisfies all conditions.
func matchesRequest(r *http.Request, cond MatchCondition) bool {
	for key, val := range cond.Header {
		if !strings.EqualFold(r.Header.Get(key), val) {
			return false
		}
	}

	q := r.URL.Query()
	for key, val := range cond.Query {
		if q.Get(key) != val {
			return false
		}
	}

	if cond.Body != "" {
		body := r.Context().Value(bodyKey{})
		if body == nil {
			return false
		}
		if !strings.Contains(body.(string), cond.Body) {
			return false
		}
	}

	return true
}

// bodyKey is a context key for storing the request body string.
type bodyKey struct{}
