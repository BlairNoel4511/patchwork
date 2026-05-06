package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/patchwork/config"
)

func makeTestConfig() *config.Config {
	return &config.Config{
		Port: 8080,
		Routes: []config.Route{
			{
				Path:   "/hello",
				Method: "GET",
				Status: 200,
				Body:   `{"message": "hello"}`,
			},
			{
				Path:   "/created",
				Method: "POST",
				Status: 201,
				Body:   `{"created": true}`,
			},
		},
	}
}

func TestNew_RoutesAreRegistered(t *testing.T) {
	cfg := makeTestConfig()
	srv := New(cfg)

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/hello")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestNew_CORSHeaderPresent(t *testing.T) {
	cfg := makeTestConfig()
	srv := New(cfg)

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/hello")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected CORS header *, got %q", got)
	}
}

func TestNew_UnknownRouteReturns404(t *testing.T) {
	cfg := makeTestConfig()
	srv := New(cfg)

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/not-found")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
