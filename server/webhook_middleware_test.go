package server

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/patchwork/config"
)

func TestWebhookMiddleware_NoWebhooks(t *testing.T) {
	route := config.Route{}
	d := NewWebhookDispatcher()

	handler := WebhookMiddleware(route, d)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestWebhookMiddleware_FiresAfterResponse(t *testing.T) {
	var called atomic.Int32
	extSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer extSrv.Close()

	route := config.Route{
		Webhooks: []config.Webhook{
			{Method: "POST", URL: extSrv.URL},
		},
	}
	d := NewWebhookDispatcher()

	handler := WebhookMiddleware(route, d)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	time.Sleep(150 * time.Millisecond)
	if called.Load() != 1 {
		t.Fatalf("expected 1 webhook call, got %d", called.Load())
	}
}
