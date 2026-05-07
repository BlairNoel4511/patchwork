package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProxyMiddleware_NoProxyURL(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	h := ProxyMiddleware("", next)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", rec.Code)
	}
}

func TestProxyMiddleware_InvalidURL(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	h := ProxyMiddleware("not-a-url", next)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	// Falls through to next because host is empty.
	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", rec.Code)
	}
}

func TestProxyMiddleware_ForwardsToUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-From-Upstream", "yes")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("upstream body"))
	}))
	defer upstream.Close()

	h := ProxyMiddleware(upstream.URL, http.NotFoundHandler())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/some/path", nil))

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
	if rec.Header().Get("X-From-Upstream") != "yes" {
		t.Error("expected upstream header to be proxied")
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "upstream body" {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestProxyMiddleware_ForwardsRequestHeaders(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Got-Auth", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	h := ProxyMiddleware(upstream.URL, http.NotFoundHandler())

	req := httptest.NewRequest(http.MethodPost, "/api", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer token123")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Header().Get("X-Got-Auth") != "Bearer token123" {
		t.Errorf("authorization header not forwarded, got: %s", rec.Header().Get("X-Got-Auth"))
	}
}

func TestSingleJoiningSlash(t *testing.T) {
	cases := []struct{ a, b, want string }{
		{"", "/foo", "/foo"},
		{"/base", "/foo", "/base/foo"},
		{"/base/", "/foo", "/base/foo"},
		{"/base", "foo", "/base/foo"},
	}
	for _, c := range cases {
		got := singleJoiningSlash(c.a, c.b)
		if got != c.want {
			t.Errorf("singleJoiningSlash(%q, %q) = %q, want %q", c.a, c.b, got, c.want)
		}
	}
}
