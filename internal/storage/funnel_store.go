package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// FunnelStore handles persistence of saved multi-step funnels in SQLite.
type FunnelStore struct {
	db *DB
}

// NewFunnelStore creates a FunnelStore instance.
func NewFunnelStore(db *DB) *FunnelStore {
	return &FunnelStore{db: db}
}

// Save inserts or updates a saved funnel configuration.
func (s *FunnelStore) Save(f *SavedFunnel) error {
	if f.ID == "" {
		bytes := make([]byte, 8)
		_, _ = rand.Read(bytes)
		f.ID = "fn_" + hex.EncodeToString(bytes)
	}
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO funnels (id, site_id, name, steps_json, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			steps_json = excluded.steps_json
	`
	_, err := s.db.Exec(query, f.ID, f.SiteID, f.Name, f.StepsJSON, f.CreatedAt.Format("2006-01-02 15:04:05"))
	return err
}

// ListBySite returns all saved funnels for a site.
func (s *FunnelStore) ListBySite(siteID string) ([]*SavedFunnel, error) {
	query := `
		SELECT id, site_id, name, steps_json, created_at
		FROM funnels
		WHERE site_id = ?
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(query, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query saved funnels: %w", err)
	}
	defer rows.Close()

	var list []*SavedFunnel
	for rows.Next() {
		var f SavedFunnel
		var rawTime interface{}
		if err := rows.Scan(&f.ID, &f.SiteID, &f.Name, &f.StepsJSON, &rawTime); err != nil {
			return nil, err
		}
		f.CreatedAt = parseTimeHelper(rawTime)
		list = append(list, &f)
	}

	return list, nil
}

// Delete removes a saved funnel by ID.
func (s *FunnelStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM funnels WHERE id = ?", id)
	return err
}
