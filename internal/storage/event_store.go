package storage

import (
	"fmt"
	"strings"
	"time"
)

// EventStore handles querying and saving event data.
type EventStore struct {
	db *DB
}

// NewEventStore creates a new EventStore instance.
func NewEventStore(db *DB) *EventStore {
	return &EventStore{db: db}
}

// InsertBatch saves a batch of events and updates associated sessions inside a single transaction.
func (s *EventStore) InsertBatch(events []*Event) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	eventStmt, err := tx.Prepare(`
		INSERT INTO events (
			site_id, session_id, visitor_id, event_name, url, path,
			referrer, referrer_domain, browser, os, device, country,
			properties_json, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare event statement: %w", err)
	}
	defer eventStmt.Close()

	sessionStmt, err := tx.Prepare(`
		INSERT INTO sessions (
			id, site_id, visitor_id, started_at, last_seen_at,
			duration_seconds, event_count, entry_path, exit_path,
			browser, os, device, country, has_replay
		) VALUES (?, ?, ?, ?, ?, 0, 1, ?, ?, ?, ?, ?, ?, 0)
		ON CONFLICT(id) DO UPDATE SET
			last_seen_at = excluded.last_seen_at,
			duration_seconds = MAX(0, CAST((strftime('%s', excluded.last_seen_at) - strftime('%s', sessions.started_at)) AS INTEGER)),
			event_count = sessions.event_count + 1,
			exit_path = excluded.exit_path
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare session statement: %w", err)
	}
	defer sessionStmt.Close()

	for _, e := range events {
		if e.CreatedAt.IsZero() {
			e.CreatedAt = time.Now().UTC()
		}
		formattedTime := e.CreatedAt.Format("2006-01-02 15:04:05")

		_, err := eventStmt.Exec(
			e.SiteID, e.SessionID, e.VisitorID, e.EventName, e.URL, e.Path,
			e.Referrer, e.ReferrerDomain, e.Browser, e.OS, e.Device, e.Country,
			e.PropertiesJSON, formattedTime,
		)
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		_, err = sessionStmt.Exec(
			e.SessionID, e.SiteID, e.VisitorID, formattedTime, formattedTime,
			e.Path, e.Path, e.Browser, e.OS, e.Device, e.Country,
		)
		if err != nil {
			return fmt.Errorf("failed to upsert session: %w", err)
		}
	}

	return tx.Commit()
}

// GetMetricsSummary calculates high-level aggregates within a timeframe.
func (s *EventStore) GetMetricsSummary(siteID string, start, end time.Time) (*SummaryMetrics, error) {
	startTime := start.UTC().Format("2006-01-02 15:04:05")
	endTime := end.UTC().Format("2006-01-02 15:04:05")

	var summary SummaryMetrics

	eventQuery := `
		SELECT
			COALESCE(COUNT(CASE WHEN event_name = '$pageview' THEN 1 END), 0) as pageviews,
			COALESCE(COUNT(DISTINCT visitor_id), 0) as unique_visitors
		FROM events
		WHERE site_id = ? AND created_at >= ? AND created_at <= ?
	`
	err := s.db.QueryRow(eventQuery, siteID, startTime, endTime).Scan(&summary.Pageviews, &summary.UniqueVisitors)
	if err != nil {
		return nil, fmt.Errorf("failed to query event summary: %w", err)
	}

	sessionQuery := `
		SELECT
			COUNT(*) as total_sessions,
			COALESCE(COUNT(CASE WHEN event_count = 1 THEN 1 END), 0) as bounced_sessions,
			COALESCE(AVG(duration_seconds), 0) as avg_duration
		FROM sessions
		WHERE site_id = ? AND started_at >= ? AND started_at <= ?
	`
	var totalSessions, bouncedSessions int64
	var avgDuration float64
	err = s.db.QueryRow(sessionQuery, siteID, startTime, endTime).Scan(&totalSessions, &bouncedSessions, &avgDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to query session summary: %w", err)
	}

	summary.Sessions = totalSessions
	summary.AvgDurationSeconds = int(avgDuration)
	if totalSessions > 0 {
		summary.BounceRate = float64(bouncedSessions) / float64(totalSessions) * 100.0
	}

	return &summary, nil
}

// GetTimeseries retrieves views and visitor counts grouped by hour or day.
func (s *EventStore) GetTimeseries(siteID string, start, end time.Time) ([]*TimeseriesPoint, error) {
	startTime := start.UTC().Format("2006-01-02 15:04:05")
	endTime := end.UTC().Format("2006-01-02 15:04:05")

	// If interval is <= 48 hours, group by hour; otherwise group by day
	isHourly := end.Sub(start) <= 48*time.Hour
	groupFormat := "%Y-%m-%d %H:00:00"
	if !isHourly {
		groupFormat = "%Y-%m-%d"
	}

	query := fmt.Sprintf(`
		SELECT
			strftime('%s', created_at) as bucket,
			COALESCE(COUNT(CASE WHEN event_name = '$pageview' THEN 1 END), 0) as views,
			COALESCE(COUNT(DISTINCT visitor_id), 0) as visitors
		FROM events
		WHERE site_id = ? AND created_at >= ? AND created_at <= ?
		GROUP BY bucket
		ORDER BY bucket ASC
	`, groupFormat)

	rows, err := s.db.Query(query, siteID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query timeseries: %w", err)
	}
	defer rows.Close()

	var points []*TimeseriesPoint
	for rows.Next() {
		var p TimeseriesPoint
		if err := rows.Scan(&p.Timestamp, &p.Views, &p.Visitors); err != nil {
			return nil, err
		}
		points = append(points, &p)
	}

	return points, nil
}

// GetTopPages returns pageview counts per path.
func (s *EventStore) GetTopPages(siteID string, start, end time.Time, limit int) ([]*PageStat, error) {
	startTime := start.UTC().Format("2006-01-02 15:04:05")
	endTime := end.UTC().Format("2006-01-02 15:04:05")

	query := `
		SELECT
			path,
			COUNT(*) as views,
			COUNT(DISTINCT visitor_id) as visitors
		FROM events
		WHERE site_id = ? AND event_name = '$pageview' AND created_at >= ? AND created_at <= ?
		GROUP BY path
		ORDER BY views DESC
		LIMIT ?
	`
	rows, err := s.db.Query(query, siteID, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top pages: %w", err)
	}
	defer rows.Close()

	var pages []*PageStat
	var totalViews int64
	for rows.Next() {
		var p PageStat
		if err := rows.Scan(&p.Path, &p.Views, &p.Visitors); err != nil {
			return nil, err
		}
		totalViews += p.Views
		pages = append(pages, &p)
	}

	if totalViews > 0 {
		for _, p := range pages {
			p.Percentage = (float64(p.Views) / float64(totalViews)) * 100.0
		}
	}

	return pages, nil
}

// GetTopReferrers returns counts per referring domain.
func (s *EventStore) GetTopReferrers(siteID string, start, end time.Time, limit int) ([]*ReferrerStat, error) {
	startTime := start.UTC().Format("2006-01-02 15:04:05")
	endTime := end.UTC().Format("2006-01-02 15:04:05")

	query := `
		SELECT
			CASE WHEN referrer_domain = '' THEN 'direct' ELSE referrer_domain END as ref,
			COUNT(*) as count
		FROM events
		WHERE site_id = ? AND created_at >= ? AND created_at <= ?
		GROUP BY ref
		ORDER BY count DESC
		LIMIT ?
	`
	rows, err := s.db.Query(query, siteID, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top referrers: %w", err)
	}
	defer rows.Close()

	var referrers []*ReferrerStat
	var totalCount int64
	for rows.Next() {
		var r ReferrerStat
		if err := rows.Scan(&r.Referrer, &r.Count); err != nil {
			return nil, err
		}
		totalCount += r.Count
		referrers = append(referrers, &r)
	}

	if totalCount > 0 {
		for _, r := range referrers {
			r.Percentage = (float64(r.Count) / float64(totalCount)) * 100.0
		}
	}

	return referrers, nil
}

// GetBreakdown calculates counts for dimensions like browser, os, device, or country.
func (s *EventStore) GetBreakdown(siteID, dimension string, start, end time.Time, limit int) ([]*BreakdownStat, error) {
	// Whitelist allowed dimensions to ensure safety
	dimCol := ""
	switch strings.ToLower(dimension) {
	case "browser":
		dimCol = "browser"
	case "os":
		dimCol = "os"
	case "device":
		dimCol = "device"
	case "country":
		dimCol = "country"
	default:
		return nil, fmt.Errorf("invalid breakdown dimension: %s", dimension)
	}

	startTime := start.UTC().Format("2006-01-02 15:04:05")
	endTime := end.UTC().Format("2006-01-02 15:04:05")

	query := fmt.Sprintf(`
		SELECT
			CASE WHEN %s = '' THEN 'unknown' ELSE %s END as dim,
			COUNT(*) as count
		FROM events
		WHERE site_id = ? AND created_at >= ? AND created_at <= ?
		GROUP BY dim
		ORDER BY count DESC
		LIMIT ?
	`, dimCol, dimCol)

	rows, err := s.db.Query(query, siteID, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query breakdown %s: %w", dimension, err)
	}
	defer rows.Close()

	var stats []*BreakdownStat
	var totalCount int64
	for rows.Next() {
		var b BreakdownStat
		if err := rows.Scan(&b.Name, &b.Count); err != nil {
			return nil, err
		}
		totalCount += b.Count
		stats = append(stats, &b)
	}

	if totalCount > 0 {
		for _, b := range stats {
			b.Percentage = (float64(b.Count) / float64(totalCount)) * 100.0
		}
	}

	return stats, nil
}

// GetRecentEvents retrieves the latest events recorded for a site.
func (s *EventStore) GetRecentEvents(siteID string, limit int) ([]*Event, error) {
	query := `
		SELECT
			id, site_id, session_id, visitor_id, event_name, url, path,
			referrer, referrer_domain, browser, os, device, country,
			properties_json, created_at
		FROM events
		WHERE site_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`
	rows, err := s.db.Query(query, siteID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent events: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		var e Event
		var timeStr string
		err := rows.Scan(
			&e.ID, &e.SiteID, &e.SessionID, &e.VisitorID, &e.EventName, &e.URL, &e.Path,
			&e.Referrer, &e.ReferrerDomain, &e.Browser, &e.OS, &e.Device, &e.Country,
			&e.PropertiesJSON, &timeStr,
		)
		if err != nil {
			return nil, err
		}
		e.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", timeStr)
		events = append(events, &e)
	}

	return events, nil
}

// GetSessions returns a paginated list of user sessions.
func (s *EventStore) GetSessions(siteID string, limit, offset int) ([]*Session, int64, error) {
	var total int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM sessions WHERE site_id = ?", siteID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	query := `
		SELECT
			id, site_id, visitor_id, started_at, last_seen_at,
			duration_seconds, event_count, entry_path, exit_path,
			browser, os, device, country, has_replay
		FROM sessions
		WHERE site_id = ?
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := s.db.Query(query, siteID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var sess Session
		var startedStr, lastSeenStr string
		err := rows.Scan(
			&sess.ID, &sess.SiteID, &sess.VisitorID, &startedStr, &lastSeenStr,
			&sess.DurationSeconds, &sess.EventCount, &sess.EntryPath, &sess.ExitPath,
			&sess.Browser, &sess.OS, &sess.Device, &sess.Country, &sess.HasReplay,
		)
		if err != nil {
			return nil, 0, err
		}
		sess.StartedAt, _ = time.Parse("2006-01-02 15:04:05", startedStr)
		sess.LastSeenAt, _ = time.Parse("2006-01-02 15:04:05", lastSeenStr)
		sessions = append(sessions, &sess)
	}

	return sessions, total, nil
}

// MarkSessionHasReplay sets the has_replay flag to true for a session.
func (s *EventStore) MarkSessionHasReplay(sessionID string) error {
	_, err := s.db.Exec("UPDATE sessions SET has_replay = 1 WHERE id = ?", sessionID)
	return err
}
