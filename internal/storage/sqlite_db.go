package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQLite database connection.
type DB struct {
	*sql.DB
}

// OpenDB opens a SQLite database with WAL mode and optimal pragmas.
func OpenDB(dbPath string) (*DB, error) {
	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create db directory: %w", err)
			}
		}
	}

	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=5000&_cache_size=-20000", dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	wrapped := &DB{DB: db}
	if err := wrapped.migrate(); err != nil {
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return wrapped, nil
}

func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sites (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		domain TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	INSERT OR IGNORE INTO sites (id, name, domain) VALUES ('default', 'Default Site', 'localhost');

	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		site_id TEXT NOT NULL DEFAULT 'default',
		session_id TEXT NOT NULL,
		visitor_id TEXT NOT NULL,
		event_name TEXT NOT NULL,
		url TEXT NOT NULL DEFAULT '',
		path TEXT NOT NULL DEFAULT '',
		referrer TEXT NOT NULL DEFAULT '',
		referrer_domain TEXT NOT NULL DEFAULT '',
		browser TEXT NOT NULL DEFAULT '',
		os TEXT NOT NULL DEFAULT '',
		device TEXT NOT NULL DEFAULT 'desktop',
		country TEXT NOT NULL DEFAULT '',
		properties_json TEXT NOT NULL DEFAULT '{}',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_events_site_time ON events(site_id, created_at);
	CREATE INDEX IF NOT EXISTS idx_events_session ON events(session_id);
	CREATE INDEX IF NOT EXISTS idx_events_visitor ON events(visitor_id);
	CREATE INDEX IF NOT EXISTS idx_events_name ON events(site_id, event_name, created_at);
	CREATE INDEX IF NOT EXISTS idx_events_path ON events(site_id, path);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL DEFAULT 'default',
		visitor_id TEXT NOT NULL,
		started_at DATETIME NOT NULL,
		last_seen_at DATETIME NOT NULL,
		duration_seconds INTEGER NOT NULL DEFAULT 0,
		event_count INTEGER NOT NULL DEFAULT 1,
		entry_path TEXT NOT NULL DEFAULT '',
		exit_path TEXT NOT NULL DEFAULT '',
		browser TEXT NOT NULL DEFAULT '',
		os TEXT NOT NULL DEFAULT '',
		device TEXT NOT NULL DEFAULT 'desktop',
		country TEXT NOT NULL DEFAULT '',
		has_replay BOOLEAN NOT NULL DEFAULT 0,
		has_rage_click BOOLEAN NOT NULL DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_site_started ON sessions(site_id, started_at DESC);
	CREATE INDEX IF NOT EXISTS idx_sessions_visitor ON sessions(visitor_id);

	CREATE TABLE IF NOT EXISTS replays (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		chunk_index INTEGER NOT NULL,
		event_count INTEGER NOT NULL DEFAULT 0,
		compressed_data BLOB NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(session_id, chunk_index)
	);

	CREATE INDEX IF NOT EXISTS idx_replays_session ON replays(session_id, chunk_index);

	CREATE TABLE IF NOT EXISTS funnels (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL DEFAULT 'default',
		name TEXT NOT NULL,
		steps_json TEXT NOT NULL DEFAULT '[]',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_funnels_site ON funnels(site_id);

	CREATE TABLE IF NOT EXISTS webhooks (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL DEFAULT 'default',
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		event_filter TEXT NOT NULL DEFAULT 'all',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_webhooks_site ON webhooks(site_id);
	`

	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// Safe alter migrations for existing installations
	_, _ = db.Exec("ALTER TABLE sessions ADD COLUMN has_rage_click BOOLEAN NOT NULL DEFAULT 0")

	return nil
}
