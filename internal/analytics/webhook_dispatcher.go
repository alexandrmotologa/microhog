package analytics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alexandrmotologa/microhog/internal/storage"
)

// WebhookDispatcher delivers real-time notifications to Slack, Discord, or generic HTTP endpoints.
type WebhookDispatcher struct {
	webhookStore *storage.WebhookStore
	client       *http.Client
}

// NewWebhookDispatcher creates a WebhookDispatcher.
func NewWebhookDispatcher(ws *storage.WebhookStore) *WebhookDispatcher {
	return &WebhookDispatcher{
		webhookStore: ws,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Dispatch checks registered webhooks and fires notifications asynchronously.
func (wd *WebhookDispatcher) Dispatch(e *storage.Event) {
	// Skip high-frequency internal events from overwhelming webhooks
	if e.EventName == "$pageview" || e.EventName == "$web_vitals" {
		return
	}

	webhooks, err := wd.webhookStore.ListBySite(e.SiteID)
	if err != nil || len(webhooks) == 0 {
		return
	}

	for _, wh := range webhooks {
		if wh.EventFilter != "all" && wh.EventFilter != e.EventName {
			continue
		}

		go wd.sendNotification(wh.URL, e)
	}
}

func (wd *WebhookDispatcher) sendNotification(targetURL string, e *storage.Event) {
	// Formatted message text suitable for Slack, Discord, and Telegram/Generic
	msgText := fmt.Sprintf("🔔 *MicroHog Alert* [%s]\nEvent: `%s` on `%s`\nVisitor: `%s` | Geo: `%s` | Device: `%s` (%s)",
		e.SiteID, e.EventName, e.Path, e.VisitorID[:8], e.Country, e.Device, e.Browser)

	payload := map[string]interface{}{
		"text":       msgText,
		"content":    msgText, // Discord fallback
		"event_name": e.EventName,
		"path":       e.Path,
		"site_id":    e.SiteID,
		"visitor_id": e.VisitorID,
		"country":    e.Country,
		"timestamp":  e.CreatedAt.Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MicroHog-Webhook/1.0")

	resp, err := wd.client.Do(req)
	if err == nil && resp != nil {
		_ = resp.Body.Close()
	}
}

// Test sends a verification ping to the webhook URL.
func (wd *WebhookDispatcher) Test(targetURL string) error {
	payload := map[string]interface{}{
		"text":    "🚀 *MicroHog Webhook Test*: connection verified successfully!",
		"content": "🚀 *MicroHog Webhook Test*: connection verified successfully!",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := wd.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("remote endpoint returned HTTP %d", resp.StatusCode)
	}

	return nil
}
