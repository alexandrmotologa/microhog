package storage

import (
	"encoding/json"
	"fmt"
	"sort"
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
			browser, os, device, country, has_replay, has_rage_click
		) VALUES (?, ?, ?, ?, ?, 0, 1, ?, ?, ?, ?, ?, ?, 0, ?)
		ON CONFLICT(id) DO UPDATE SET
			last_seen_at = excluded.last_seen_at,
			duration_seconds = MAX(0, CAST((strftime('%s', excluded.last_seen_at) - strftime('%s', sessions.started_at)) AS INTEGER)),
			event_count = sessions.event_count + 1,
			exit_path = excluded.exit_path,
			has_rage_click = CASE WHEN excluded.has_rage_click = 1 THEN 1 ELSE sessions.has_rage_click END
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

		isRageClick := 0
		if e.EventName == "$rage_click" || strings.Contains(e.PropertiesJSON, `"rage_click":true`) {
			isRageClick = 1
		}

		_, err = sessionStmt.Exec(
			e.SessionID, e.SiteID, e.VisitorID, formattedTime, formattedTime,
			e.Path, e.Path, e.Browser, e.OS, e.Device, e.Country, isRageClick,
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
		var rawTime interface{}
		err := rows.Scan(
			&e.ID, &e.SiteID, &e.SessionID, &e.VisitorID, &e.EventName, &e.URL, &e.Path,
			&e.Referrer, &e.ReferrerDomain, &e.Browser, &e.OS, &e.Device, &e.Country,
			&e.PropertiesJSON, &rawTime,
		)
		if err != nil {
			return nil, err
		}
		e.CreatedAt = parseTimeHelper(rawTime)
		events = append(events, &e)
	}

	return events, nil
}

// SearchEvents searches events by name, path, or properties JSON.
func (s *EventStore) SearchEvents(siteID, queryStr, eventName string, limit, offset int) ([]*Event, int64, error) {
	where := []string{"site_id = ?"}
	args := []interface{}{siteID}

	if eventName != "" && eventName != "all" {
		where = append(where, "event_name = ?")
		args = append(args, eventName)
	}

	if queryStr != "" {
		where = append(where, "(path LIKE ? OR properties_json LIKE ? OR visitor_id LIKE ?)")
		pattern := "%" + queryStr + "%"
		args = append(args, pattern, pattern, pattern)
	}

	whereClause := strings.Join(where, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM events WHERE %s", whereClause)
	var total int64
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT
			id, site_id, session_id, visitor_id, event_name, url, path,
			referrer, referrer_domain, browser, os, device, country,
			properties_json, created_at
		FROM events
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	argsWithPaging := append(args, limit, offset)
	rows, err := s.db.Query(query, argsWithPaging...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		var e Event
		var rawTime interface{}
		err := rows.Scan(
			&e.ID, &e.SiteID, &e.SessionID, &e.VisitorID, &e.EventName, &e.URL, &e.Path,
			&e.Referrer, &e.ReferrerDomain, &e.Browser, &e.OS, &e.Device, &e.Country,
			&e.PropertiesJSON, &rawTime,
		)
		if err != nil {
			return nil, 0, err
		}
		e.CreatedAt = parseTimeHelper(rawTime)
		events = append(events, &e)
	}

	return events, total, nil
}

// GetSessions returns a paginated list of user sessions, with optional rage click filter.
func (s *EventStore) GetSessions(siteID string, limit, offset int, rageOnly bool) ([]*Session, int64, error) {
	where := "site_id = ?"
	if rageOnly {
		where += " AND has_rage_click = 1"
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM sessions WHERE %s", where)
	err := s.db.QueryRow(countQuery, siteID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT
			id, site_id, visitor_id, started_at, last_seen_at,
			duration_seconds, event_count, entry_path, exit_path,
			browser, os, device, country, has_replay, has_rage_click
		FROM sessions
		WHERE %s
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`, where)

	rows, err := s.db.Query(query, siteID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var sess Session
		var rawStarted, rawLastSeen interface{}
		err := rows.Scan(
			&sess.ID, &sess.SiteID, &sess.VisitorID, &rawStarted, &rawLastSeen,
			&sess.DurationSeconds, &sess.EventCount, &sess.EntryPath, &sess.ExitPath,
			&sess.Browser, &sess.OS, &sess.Device, &sess.Country, &sess.HasReplay, &sess.HasRageClick,
		)
		if err != nil {
			return nil, 0, err
		}
		sess.StartedAt = parseTimeHelper(rawStarted)
		sess.LastSeenAt = parseTimeHelper(rawLastSeen)
		sessions = append(sessions, &sess)
	}

	return sessions, total, nil
}

// GetWebVitals calculates Core Web Vitals (LCP, INP, CLS p75) and breaks down scores per path.
func (s *EventStore) GetWebVitals(siteID string, start, end time.Time) (*WebVitalsReport, error) {
	startTime := start.UTC().Format("2006-01-02 15:04:05")
	endTime := end.UTC().Format("2006-01-02 15:04:05")

	query := `
		SELECT path, properties_json
		FROM events
		WHERE site_id = ? AND event_name = '$web_vitals' AND created_at >= ? AND created_at <= ?
	`
	rows, err := s.db.Query(query, siteID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query web vitals: %w", err)
	}
	defer rows.Close()

	var allLCP []float64
	var allINP []float64
	var allCLS []float64

	type pageAccumulator struct {
		lcpSum float64
		inpSum float64
		clsSum float64
		count  int64
	}
	pageMap := make(map[string]*pageAccumulator)

	for rows.Next() {
		var path, propsJSON string
		if err := rows.Scan(&path, &propsJSON); err != nil {
			continue
		}

		var props struct {
			LCP float64 `json:"lcp"`
			INP float64 `json:"inp"`
			CLS float64 `json:"cls"`
		}
		if err := json.Unmarshal([]byte(propsJSON), &props); err != nil {
			continue
		}

		acc, ok := pageMap[path]
		if !ok {
			acc = &pageAccumulator{}
			pageMap[path] = acc
		}
		acc.count++

		if props.LCP > 0 {
			allLCP = append(allLCP, props.LCP)
			acc.lcpSum += props.LCP
		}
		if props.INP > 0 {
			allINP = append(allINP, props.INP)
			acc.inpSum += props.INP
		}
		if props.CLS >= 0 {
			allCLS = append(allCLS, props.CLS)
			acc.clsSum += props.CLS
		}
	}

	computeMetric := func(name string, vals []float64, goodThresh, poorThresh float64) VitalMetric {
		if len(vals) == 0 {
			return VitalMetric{Name: name, P75: 0, Median: 0, Rating: "good", Count: 0}
		}
		sort.Float64s(vals)
		p75Idx := int(float64(len(vals)) * 0.75)
		if p75Idx >= len(vals) {
			p75Idx = len(vals) - 1
		}
		medianIdx := len(vals) / 2

		p75 := vals[p75Idx]
		median := vals[medianIdx]

		rating := "good"
		if p75 > poorThresh {
			rating = "poor"
		} else if p75 > goodThresh {
			rating = "needs_improvement"
		}

		return VitalMetric{
			Name:   name,
			P75:    p75,
			Median: median,
			Rating: rating,
			Count:  int64(len(vals)),
		}
	}

	// LCP: Good <= 2500ms, Poor > 4000ms
	lcpMetric := computeMetric("LCP", allLCP, 2500, 4000)
	// INP: Good <= 200ms, Poor > 500ms
	inpMetric := computeMetric("INP", allINP, 200, 500)
	// CLS: Good <= 0.1, Poor > 0.25
	clsMetric := computeMetric("CLS", allCLS, 0.1, 0.25)

	var pages []*PageVitals
	for path, acc := range pageMap {
		if acc.count == 0 {
			continue
		}
		pages = append(pages, &PageVitals{
			Path:  path,
			LCP:   acc.lcpSum / float64(acc.count),
			INP:   acc.inpSum / float64(acc.count),
			CLS:   acc.clsSum / float64(acc.count),
			Count: acc.count,
		})
	}
	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Count > pages[j].Count
	})

	return &WebVitalsReport{
		LCP:   lcpMetric,
		INP:   inpMetric,
		CLS:   clsMetric,
		Pages: pages,
	}, nil
}

// ExportEvents retrieves all events in a timeframe for CSV or JSON download.
func (s *EventStore) ExportEvents(siteID string, start, end time.Time) ([]*Event, error) {
	startTime := start.UTC().Format("2006-01-02 15:04:05")
	endTime := end.UTC().Format("2006-01-02 15:04:05")

	query := `
		SELECT
			id, site_id, session_id, visitor_id, event_name, url, path,
			referrer, referrer_domain, browser, os, device, country,
			properties_json, created_at
		FROM events
		WHERE site_id = ? AND created_at >= ? AND created_at <= ?
		ORDER BY created_at ASC
	`
	rows, err := s.db.Query(query, siteID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to export events: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		var e Event
		var rawTime interface{}
		err := rows.Scan(
			&e.ID, &e.SiteID, &e.SessionID, &e.VisitorID, &e.EventName, &e.URL, &e.Path,
			&e.Referrer, &e.ReferrerDomain, &e.Browser, &e.OS, &e.Device, &e.Country,
			&e.PropertiesJSON, &rawTime,
		)
		if err != nil {
			return nil, err
		}
		e.CreatedAt = parseTimeHelper(rawTime)
		events = append(events, &e)
	}

	return events, nil
}

// ExportSessions retrieves all sessions in a timeframe for CSV or JSON download.
func (s *EventStore) ExportSessions(siteID string, start, end time.Time) ([]*Session, error) {
	startTime := start.UTC().Format("2006-01-02 15:04:05")
	endTime := end.UTC().Format("2006-01-02 15:04:05")

	query := `
		SELECT
			id, site_id, visitor_id, started_at, last_seen_at,
			duration_seconds, event_count, entry_path, exit_path,
			browser, os, device, country, has_replay, has_rage_click
		FROM sessions
		WHERE site_id = ? AND started_at >= ? AND started_at <= ?
		ORDER BY started_at ASC
	`
	rows, err := s.db.Query(query, siteID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to export sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var sess Session
		var rawStarted, rawLastSeen interface{}
		err := rows.Scan(
			&sess.ID, &sess.SiteID, &sess.VisitorID, &rawStarted, &rawLastSeen,
			&sess.DurationSeconds, &sess.EventCount, &sess.EntryPath, &sess.ExitPath,
			&sess.Browser, &sess.OS, &sess.Device, &sess.Country, &sess.HasReplay, &sess.HasRageClick,
		)
		if err != nil {
			return nil, err
		}
		sess.StartedAt = parseTimeHelper(rawStarted)
		sess.LastSeenAt = parseTimeHelper(rawLastSeen)
		sessions = append(sessions, &sess)
	}

	return sessions, nil
}

// MarkSessionHasReplay sets the has_replay flag to true for a session.
func (s *EventStore) MarkSessionHasReplay(sessionID string) error {
	_, err := s.db.Exec("UPDATE sessions SET has_replay = 1 WHERE id = ?", sessionID)
	return err
}

func parseTimeHelper(val interface{}) time.Time {
	switch v := val.(type) {
	case time.Time:
		return v
	case string:
		for _, layout := range []string{
			"2006-01-02 15:04:05",
			time.RFC3339,
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05 -0700 MST",
			"2006-01-02 15:04:05 +0000 UTC",
		} {
			if t, err := time.Parse(layout, v); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}
