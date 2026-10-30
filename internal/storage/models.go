package storage

import "time"

// Event represents an individual recorded product analytics event.
type Event struct {
	ID             int64     `json:"id"`
	SiteID         string    `json:"site_id"`
	SessionID      string    `json:"session_id"`
	VisitorID      string    `json:"visitor_id"`
	EventName      string    `json:"event_name"`
	URL            string    `json:"url"`
	Path           string    `json:"path"`
	Referrer       string    `json:"referrer"`
	ReferrerDomain string    `json:"referrer_domain"`
	Browser        string    `json:"browser"`
	OS             string    `json:"os"`
	Device         string    `json:"device"`
	Country        string    `json:"country"`
	PropertiesJSON string    `json:"properties_json"`
	CreatedAt      time.Time `json:"created_at"`
}

// Session represents a single user interaction visit.
type Session struct {
	ID              string    `json:"id"`
	SiteID          string    `json:"site_id"`
	VisitorID       string    `json:"visitor_id"`
	StartedAt       time.Time `json:"started_at"`
	LastSeenAt      time.Time `json:"last_seen_at"`
	DurationSeconds int       `json:"duration_seconds"`
	EventCount      int       `json:"event_count"`
	EntryPath       string    `json:"entry_path"`
	ExitPath        string    `json:"exit_path"`
	Browser         string    `json:"browser"`
	OS              string    `json:"os"`
	Device          string    `json:"device"`
	Country         string    `json:"country"`
	HasReplay       bool      `json:"has_replay"`
}

// ReplayChunk represents a compressed chunk of DOM mutation records.
type ReplayChunk struct {
	ID             int64     `json:"id"`
	SessionID      string    `json:"session_id"`
	ChunkIndex     int       `json:"chunk_index"`
	EventCount     int       `json:"event_count"`
	CompressedData []byte    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
}

// SummaryMetrics holds aggregated high-level KPIs.
type SummaryMetrics struct {
	Pageviews          int64   `json:"pageviews"`
	UniqueVisitors     int64   `json:"unique_visitors"`
	Sessions           int64   `json:"sessions"`
	BounceRate         float64 `json:"bounce_rate"`
	AvgDurationSeconds int     `json:"avg_duration_seconds"`
}

// TimeseriesPoint represents views and visitors for a specific time bucket.
type TimeseriesPoint struct {
	Timestamp string `json:"timestamp"`
	Views     int64  `json:"views"`
	Visitors  int64  `json:"visitors"`
}

// BreakdownStat represents a counted category (e.g. browser, country).
type BreakdownStat struct {
	Name       string  `json:"name"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// PageStat represents visit counts for an application route or path.
type PageStat struct {
	Path       string  `json:"path"`
	Views      int64   `json:"views"`
	Visitors   int64   `json:"visitors"`
	Percentage float64 `json:"percentage"`
}

// ReferrerStat represents referrer domain traffic statistics.
type ReferrerStat struct {
	Referrer   string  `json:"referrer"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}
