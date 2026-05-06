package server

import (
	"net/http"
	"strings"

	"github.com/user/patchwork/config"
)

const defaultContentType = "application/json"

// NewRouteHandler returns an http.HandlerFunc for the given route definition.
// The response body supports Go template syntax; see template.go for available
// template data fields (.Query, .Header, .Path, .Method).
func NewRouteHandler(route config.Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Determine content type
		contentType := defaultContentType
		for k, v := range route.Headers {
			if strings.EqualFold(k, "content-type") {
				contentType = v
			}
		}

		// Apply custom headers
		for k, v := range route.Headers {
			w.Header().Set(k, v)
		}

		// Ensure Content-Type is always set
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", contentType)
		}

		// Render body template
		body := renderTemplate(route.Body, r)

		status := route.Status
		if status == 0 {
			status = http.StatusOK
		}

		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}
