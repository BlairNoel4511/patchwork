package server

import (
	"net/http"

	"github.com/user/patchwork/config"
)

// ConditionalResponse pairs a match condition with a response definition.
type ConditionalResponse struct {
	Match    MatchCondition  `yaml:"match"`
	Response config.Response `yaml:"response"`
}

// NewConditionalHandler returns an http.Handler that evaluates a list of
// conditional responses in order, falling back to defaultResp if none match.
func NewConditionalHandler(conditions []ConditionalResponse, defaultResp config.Response) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, c := range conditions {
			if matchesRequest(r, c.Match) {
				writeResponse(w, r, c.Response)
				return
			}
		}
		writeResponse(w, r, defaultResp)
	})
}

func writeResponse(w http.ResponseWriter, r *http.Request, resp config.Response) {
	contentType := resp.ContentType
	if contentType == "" {
		contentType = "application/json"
	}
	for key, val := range resp.Headers {
		w.Header().Set(key, val)
	}
	w.Header().Set("Content-Type", contentType)

	status := resp.Status
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)

	body, err := renderJSONTemplate(resp.Body, r)
	if err != nil {
		body = resp.Body
	}
	_, _ = w.Write([]byte(body))
}
