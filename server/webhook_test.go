package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/patchwork/config"
)

func TestWebhookDispatcher_SendsRequest(t *testing.T) {
	var called atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := NewWebhookDispatcher()
	hooks := []config.Webhook{
		{Method: "POST", URL: srv.URL + "/hook", Body: `{"ok":true}`},
	}
	d.Dispatch(context.Background(), hooks)

	time.Sleep(100 * time.Millisecond)
	if called.Load() != 1 {
		t.Fatalf("expected 1 webhook call, got %d", called.Load())
	}
}

func TestWebhookDispatcher_MultipleHooks(t *testing.T) {
	var called atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := NewWebhookDispatcher()
	hooks := []config.Webhook{
		{Method: "POST", URL: srv.URL},
		{Method: "GET", URL: srv.URL},
	}
	d.Dispatch(context.Background(), hooks)

	time.Sleep(150 * time.Millisecond)
	if called.Load() != 2 {
		t.Fatalf("expected 2 webhook calls, got %d", called.Load())
	}
}

func TestWebhookDispatcher_DelayRespected(t *testing.T) {
	var called atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := NewWebhookDispatcher()
	hooks := []config.Webhook{
		{Method: "GET", URL: srv.URL, DelayMs: 80},
	}
	d.Dispatch(context.Background(), hooks)

	time.Sleep(30 * time.Millisecond)
	if called.Load() != 0 {
		t.Fatal("webhook fired before delay elapsed")
	}

	time.Sleep(120 * time.Millisecond)
	if called.Load() != 1 {
		t.Fatalf("expected 1 call after delay, got %d", called.Load())
	}
}

func TestWebhookDispatcher_ContextCancelled(t *testing.T) {
	var called atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	d := NewWebhookDispatcher()
	hooks := []config.Webhook{
		{Method: "GET", URL: srv.URL, DelayMs: 200},
	}
	d.Dispatch(ctx, hooks)
	cancel()

	time.Sleep(300 * time.Millisecond)
	if called.Load() != 0 {
		t.Fatal("webhook should not fire after context cancel")
	}
}
