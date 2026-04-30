package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/user/patchwork/config"
)

// NewRouteHandler returns an http.HandlerFunc for the given route definition.
func NewRouteHandler(route config.Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if route.Response.Delay > 0 {
			time.Sleep(time.Duration(route.Response.Delay) * time.Millisecond)
		}

		for key, value := range route.Response.Headers {
			w.Header().Set(key, value)
		}

		if _, ok := route.Response.Headers["Content-Type"]; !ok {
			w.Header().Set("Content-Type", "application/json")
		}

		w.WriteHeader(route.Response.Status)

		if route.Response.Body != nil {
			if err := json.NewEncoder(w).Encode(route.Response.Body); err != nil {
				log.Printf("error encoding response body for %s %s: %v", route.Method, route.Path, err)
			}
		}
	}
}
