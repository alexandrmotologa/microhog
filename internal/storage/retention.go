package storage

import (
	"context"
	"fmt"
	"log"
	"time"
)

// RetentionManager handles pruning historical analytics records.
type RetentionManager struct {
	db            *DB
	retentionDays int
}

// NewRetentionManager creates a RetentionManager instance.
func NewRetentionManager(db *DB, retentionDays int) *RetentionManager {
	return &RetentionManager{
		db:            db,
		retentionDays: retentionDays,
	}
}

// Prune removes events, sessions, and replays older than the given number of days.
func (r *RetentionManager) Prune(days int) (int64, error) {
	if days <= 0 {
		days = r.retentionDays
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")

	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin prune transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Delete old events
	res, err := tx.Exec("DELETE FROM events WHERE created_at < ?", cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired events: %w", err)
	}
	deletedEvents, _ := res.RowsAffected()

	// 2. Delete old replays
	_, err = tx.Exec("DELETE FROM replays WHERE created_at < ?", cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired replays: %w", err)
	}

	// 3. Delete old sessions
	_, err = tx.Exec("DELETE FROM sessions WHERE started_at < ?", cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit prune transaction: %w", err)
	}

	return deletedEvents, nil
}

// StartWorker periodically runs retention pruning once every 24 hours.
func (r *RetentionManager) StartWorker(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			deleted, err := r.Prune(r.retentionDays)
			if err != nil {
				log.Printf("[Retention] Pruning failed: %v", err)
			} else if deleted > 0 {
				log.Printf("[Retention] Pruned %d expired events", deleted)
			}
		}
	}
}
