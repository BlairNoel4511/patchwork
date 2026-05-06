package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/patchwork/config"
)

func makeRoute(strategy string, responses []config.Response) config.Route {
	return config.Route{
		Method:           "GET",
		Path:             "/test",
		ResponseStrategy: strategy,
		Responses:        responses,
	}
}

func TestSelectResponse_SingleTopLevel(t *testing.T) {
	route := config.Route{
		Method:  "GET",
		Path:    "/ping",
		Status:  200,
		Body:    `{"ok":true}`,
		Headers: map[string]string{"X-Foo": "bar"},
	}
	resp := selectResponse(route, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if resp.Status != 200 {
		t.Errorf("expected status 200, got %d", resp.Status)
	}
	if resp.Body != `{"ok":true}` {
		t.Errorf("unexpected body: %s", resp.Body)
	}
}

func TestSelectResponse_SingleInList(t *testing.T) {
	route := makeRoute("", []config.Response{
		{Status: 201, Body: "created"},
	})
	resp := selectResponse(route, httptest.NewRequest(http.MethodGet, "/test", nil))
	if resp.Status != 201 {
		t.Errorf("expected 201, got %d", resp.Status)
	}
}

func TestSelectResponse_SequentialViaHeader(t *testing.T) {
	route := makeRoute("sequential", []config.Response{
		{Status: 200, Body: "first"},
		{Status: 503, Body: "second"},
		{Status: 200, Body: "third"},
	})

	cases := []struct {
		header string
		wantStatus int
	}{
		{"0", 200},
		{"1", 503},
		{"2", 200},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Patchwork-Seq", tc.header)
		resp := selectResponse(route, req)
		if resp.Status != tc.wantStatus {
			t.Errorf("seq=%s: expected %d, got %d", tc.header, tc.wantStatus, resp.Status)
		}
	}
}

func TestSelectResponse_RandomReturnsValidEntry(t *testing.T) {
	route := makeRoute("random", []config.Response{
		{Status: 200, Body: "a"},
		{Status: 404, Body: "b"},
		{Status: 500, Body: "c"},
	})
	valid := map[int]bool{200: true, 404: true, 500: true}
	for i := 0; i < 20; i++ {
		resp := selectResponse(route, httptest.NewRequest(http.MethodGet, "/test", nil))
		if !valid[resp.Status] {
			t.Errorf("unexpected status from random selection: %d", resp.Status)
		}
	}
}

func TestSequentialIndex_OutOfBounds(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Patchwork-Seq", "99")
	idx := sequentialIndex(req, 3)
	if idx != 0 {
		t.Errorf("expected clamped index 0, got %d", idx)
	}
}
