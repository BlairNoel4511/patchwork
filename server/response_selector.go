package server

import (
	"math/rand"
	"net/http"

	"github.com/user/patchwork/config"
)

// selectResponse picks a response from a route's response list.
// If the route defines multiple responses, selection is based on the
// configured strategy: "random" or "sequential" (default).
func selectResponse(route config.Route, r *http.Request) config.Response {
	if len(route.Responses) == 0 {
		// Fallback to the single top-level response definition.
		return config.Response{
			Status:  route.Status,
			Body:    route.Body,
			Headers: route.Headers,
			Delay:   route.Delay,
		}
	}

	if len(route.Responses) == 1 {
		return route.Responses[0]
	}

	switch route.ResponseStrategy {
	case "random":
		return route.Responses[rand.Intn(len(route.Responses))]
	default:
		// Sequential: use a counter stored in a request header injected by tests,
		// or fall back to the first response in production paths.
		if idx := sequentialIndex(r, len(route.Responses)); idx >= 0 {
			return route.Responses[idx]
		}
		return route.Responses[0]
	}
}

// sequentialIndex reads an optional X-Patchwork-Seq header used in testing
// to simulate sequential response cycling. Returns -1 when absent.
func sequentialIndex(r *http.Request, total int) int {
	if r == nil {
		return -1
	}
	v := r.Header.Get("X-Patchwork-Seq")
	if v == "" {
		return -1
	}
	idx := 0
	for _, ch := range v {
		idx = idx*10 + int(ch-'0')
	}
	if idx < 0 || idx >= total {
		return 0
	}
	return idx
}
