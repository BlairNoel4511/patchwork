package server

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/user/patchwork/config"
)

// WebhookDispatcher sends outbound HTTP requests after a route is matched.
type WebhookDispatcher struct {
	client *http.Client
}

// NewWebhookDispatcher creates a dispatcher with a sensible default timeout.
func NewWebhookDispatcher() *WebhookDispatcher {
	return &WebhookDispatcher{
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// Dispatch sends all webhooks defined on the route. Errors are logged but
// never propagate to the caller so the mock response is always returned.
func (d *WebhookDispatcher) Dispatch(ctx context.Context, hooks []config.Webhook) {
	for _, hook := range hooks {
		go d.send(ctx, hook)
	}
}

func (d *WebhookDispatcher) send(ctx context.Context, hook config.Webhook) {
	delay := time.Duration(hook.DelayMs) * time.Millisecond
	if delay > 0 {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return
		}
	}

	body := bytes.NewBufferString(hook.Body)
	req, err := http.NewRequestWithContext(ctx, hook.Method, hook.URL, body)
	if err != nil {
		fmt.Printf("[webhook] build request error: %v\n", err)
		return
	}
	for k, v := range hook.Headers {
		req.Header.Set(k, v)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		fmt.Printf("[webhook] dispatch error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("[webhook] %s %s -> %d\n", hook.Method, hook.URL, resp.StatusCode)
}
