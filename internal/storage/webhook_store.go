package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// WebhookStore handles persistence of outbound notification webhooks.
type WebhookStore struct {
	db *DB
}

// NewWebhookStore creates a WebhookStore instance.
func NewWebhookStore(db *DB) *WebhookStore {
	return &WebhookStore{db: db}
}

// Save inserts or updates a webhook subscription.
func (s *WebhookStore) Save(w *Webhook) error {
	if w.ID == "" {
		bytes := make([]byte, 8)
		_, _ = rand.Read(bytes)
		w.ID = "wh_" + hex.EncodeToString(bytes)
	}
	if w.CreatedAt.IsZero() {
		w.CreatedAt = time.Now().UTC()
	}
	if w.EventFilter == "" {
		w.EventFilter = "all"
	}

	query := `
		INSERT INTO webhooks (id, site_id, name, url, event_filter, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			url = excluded.url,
			event_filter = excluded.event_filter
	`
	_, err := s.db.Exec(query, w.ID, w.SiteID, w.Name, w.URL, w.EventFilter, w.CreatedAt.Format("2006-01-02 15:04:05"))
	return err
}

// ListBySite returns all webhooks registered for a site.
func (s *WebhookStore) ListBySite(siteID string) ([]*Webhook, error) {
	query := `
		SELECT id, site_id, name, url, event_filter, created_at
		FROM webhooks
		WHERE site_id = ?
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(query, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query webhooks: %w", err)
	}
	defer rows.Close()

	var list []*Webhook
	for rows.Next() {
		var w Webhook
		var rawTime interface{}
		if err := rows.Scan(&w.ID, &w.SiteID, &w.Name, &w.URL, &w.EventFilter, &rawTime); err != nil {
			return nil, err
		}
		w.CreatedAt = parseTimeHelper(rawTime)
		list = append(list, &w)
	}

	return list, nil
}

// Delete removes a webhook by ID.
func (s *WebhookStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM webhooks WHERE id = ?", id)
	return err
}
