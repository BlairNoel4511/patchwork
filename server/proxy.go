package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ProxyMiddleware forwards the request to a upstream target when the route
// defines a proxy URL, then writes the upstream response back to the client.
func ProxyMiddleware(proxyURL string, next http.Handler) http.Handler {
	if proxyURL == "" {
		return next
	}

	target, err := url.Parse(proxyURL)
	if err != nil || target.Host == "" {
		// Invalid URL — fall through to normal handler.
		return next
	}

	client := &http.Client{Timeout: 30 * time.Second}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstream := buildUpstreamRequest(r, target)
		if upstream == nil {
			http.Error(w, "proxy: failed to build upstream request", http.StatusBadGateway)
			return
		}

		resp, err := client.Do(upstream)
		if err != nil {
			http.Error(w, fmt.Sprintf("proxy: upstream error: %v", err), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for key, vals := range resp.Header {
			for _, v := range vals {
				w.Header().Add(key, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	})
}

func buildUpstreamRequest(r *http.Request, target *url.URL) *http.Request {
	upstreamURL := *target
	upstreamURL.Path = singleJoiningSlash(target.Path, r.URL.Path)
	upstreamURL.RawQuery = r.URL.RawQuery

	req, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL.String(), r.Body)
	if err != nil {
		return nil
	}
	for key, vals := range r.Header {
		for _, v := range vals {
			req.Header.Add(key, v)
		}
	}
	req.Header.Set("X-Forwarded-Host", r.Host)
	return req
}

func singleJoiningSlash(a, b string) string {
	if a == "" {
		return b
	}
	aSlash := len(a) > 0 && a[len(a)-1] == '/'
	bSlash := len(b) > 0 && b[0] == '/'
	switch {
	case aSlash && bSlash:
		return a + b[1:]
	case !aSlash && !bSlash:
		return a + "/" + b
	}
	return a + b
}
