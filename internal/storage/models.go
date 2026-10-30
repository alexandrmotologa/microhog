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
	HasRageClick    bool      `json:"has_rage_click"`
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

// SavedFunnel represents a persisted conversion funnel definition in SQLite.
type SavedFunnel struct {
	ID        string    `json:"id"`
	SiteID    string    `json:"site_id"`
	Name      string    `json:"name"`
	StepsJSON string    `json:"steps_json"`
	CreatedAt time.Time `json:"created_at"`
}

// Webhook represents an automated outbound notification configuration.
type Webhook struct {
	ID          string    `json:"id"`
	SiteID      string    `json:"site_id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	EventFilter string    `json:"event_filter"` // "all" or specific event name like "signup_completed"
	CreatedAt   time.Time `json:"created_at"`
}

// VitalMetric aggregates scores for a single Web Vital.
type VitalMetric struct {
	Name   string  `json:"name"` // "LCP", "INP", "CLS"
	P75    float64 `json:"p75"`  // 75th percentile value
	Median float64 `json:"median"`
	Rating string  `json:"rating"` // "good", "needs_improvement", "poor"
	Count  int64   `json:"count"`
}

// PageVitals aggregates Web Vitals for an individual path.
type PageVitals struct {
	Path  string  `json:"path"`
	LCP   float64 `json:"lcp"`
	INP   float64 `json:"inp"`
	CLS   float64 `json:"cls"`
	Count int64   `json:"count"`
}

// WebVitalsReport summarizes site-wide and per-page Core Web Vitals.
type WebVitalsReport struct {
	LCP   VitalMetric   `json:"lcp"`
	INP   VitalMetric   `json:"inp"`
	CLS   VitalMetric   `json:"cls"`
	Pages []*PageVitals `json:"pages"`
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

// BreakdownStat represents a counted category.
type BreakdownStat struct {
	Name       string  `json:"name"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// PageStat represents visit counts for an application route.
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
