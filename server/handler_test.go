package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/patchwork/config"
)

func TestNewRouteHandler_StatusAndBody(t *testing.T) {
	route := config.Route{
		Method: http.MethodGet,
		Path:   "/hello",
		Response: config.Response{
			Status: http.StatusOK,
			Body:   map[string]any{"message": "hello"},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rr := httptest.NewRecorder()
	NewRouteHandler(route)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["message"] != "hello" {
		t.Errorf("expected message 'hello', got %v", body["message"])
	}
}

func TestNewRouteHandler_CustomHeaders(t *testing.T) {
	route := config.Route{
		Method: http.MethodGet,
		Path:   "/custom",
		Response: config.Response{
			Status:  http.StatusAccepted,
			Headers: map[string]string{"X-Custom-Header": "patchwork"},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/custom", nil)
	rr := httptest.NewRecorder()
	NewRouteHandler(route)(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", rr.Code)
	}
	if got := rr.Header().Get("X-Custom-Header"); got != "patchwork" {
		t.Errorf("expected header value 'patchwork', got %q", got)
	}
}

func TestNewRouteHandler_DefaultContentType(t *testing.T) {
	route := config.Route{
		Method: http.MethodGet,
		Path:   "/ct",
		Response: config.Response{
			Status: http.StatusOK,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/ct", nil)
	rr := httptest.NewRecorder()
	NewRouteHandler(route)(rr, req)

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected default Content-Type 'application/json', got %q", ct)
	}
}
