package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexandrmotologa/microhog/internal/analytics"
	"github.com/alexandrmotologa/microhog/internal/storage"
)

// APIHandler bundles handlers for dashboard REST endpoints.
type APIHandler struct {
	db                *storage.DB
	eventStore        *storage.EventStore
	replayStore       *storage.ReplayStore
	funnelStore       *storage.FunnelStore
	webhookStore      *storage.WebhookStore
	webhookDispatcher *analytics.WebhookDispatcher
	metricsService    *analytics.MetricsService
	funnelEngine      *analytics.FunnelEngine
	realtimeTracker   *analytics.RealtimeTracker
	seeder            *analytics.Seeder
	startTime         time.Time
}

// NewAPIHandler creates an APIHandler.
func NewAPIHandler(
	db *storage.DB,
	es *storage.EventStore,
	rs *storage.ReplayStore,
	fs *storage.FunnelStore,
	ws *storage.WebhookStore,
	wd *analytics.WebhookDispatcher,
	ms *analytics.MetricsService,
	fe *analytics.FunnelEngine,
	rt *analytics.RealtimeTracker,
	seeder *analytics.Seeder,
) *APIHandler {
	return &APIHandler{
		db:                db,
		eventStore:        es,
		replayStore:       rs,
		funnelStore:       fs,
		webhookStore:      ws,
		webhookDispatcher: wd,
		metricsService:    ms,
		funnelEngine:      fe,
		realtimeTracker:   rt,
		seeder:            seeder,
		startTime:         time.Now(),
	}
}

// HandleHealth provides uptime and status information.
func (a *APIHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"uptime": time.Since(a.startTime).String(),
		"server": "MicroHog/1.0",
	})
}

// HandleMetrics returns summary stats and time-series metrics.
func (a *APIHandler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	siteID := r.URL.Query().Get("site_id")
	if siteID == "" {
		siteID = "default"
	}
	timeRange := r.URL.Query().Get("range")
	if timeRange == "" {
		timeRange = "7d"
	}

	payload, err := a.metricsService.GetOverview(siteID, timeRange)
	if err != nil {
		http.Error(w, "Failed to compute metrics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

// HandleRealtime returns active visitor statistics for the past 5 minutes.
func (a *APIHandler) HandleRealtime(w http.ResponseWriter, r *http.Request) {
	siteID := r.URL.Query().Get("site_id")
	if siteID == "" {
		siteID = "default"
	}

	stats := a.realtimeTracker.GetStats(siteID)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}

// HandleFunnels calculates conversion rates for specified funnel steps.
func (a *APIHandler) HandleFunnels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SiteID string                    `json:"site_id"`
		Steps  []analytics.FunnelStepDef `json:"steps"`
		Range  string                    `json:"range"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.SiteID == "" {
		req.SiteID = "default"
	}

	now := time.Now().UTC()
	start := now.Add(-7 * 24 * time.Hour)
	if req.Range == "24h" {
		start = now.Add(-24 * time.Hour)
	} else if req.Range == "30d" {
		start = now.Add(-30 * 24 * time.Hour)
	}

	result, err := a.funnelEngine.Calculate(req.SiteID, req.Steps, start, now)
	if err != nil {
		http.Error(w, "Funnel computation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// HandleSavedFunnels handles listing and saving persisted funnels.
func (a *APIHandler) HandleSavedFunnels(w http.ResponseWriter, r *http.Request) {
	siteID := r.URL.Query().Get("site_id")
	if siteID == "" {
		siteID = "default"
	}

	switch r.Method {
	case http.MethodGet:
		funnels, err := a.funnelStore.ListBySite(siteID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(funnels)

	case http.MethodPost:
		var req struct {
			ID        string                    `json:"id"`
			SiteID    string                    `json:"site_id"`
			Name      string                    `json:"name"`
			Steps     []analytics.FunnelStepDef `json:"steps"`
			StepsJSON string                    `json:"steps_json"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if req.SiteID == "" {
			req.SiteID = siteID
		}
		if req.StepsJSON == "" && len(req.Steps) > 0 {
			b, _ := json.Marshal(req.Steps)
			req.StepsJSON = string(b)
		}
		if req.Name == "" {
			http.Error(w, "Funnel name required", http.StatusBadRequest)
			return
		}

		item := &storage.SavedFunnel{
			ID:        req.ID,
			SiteID:    req.SiteID,
			Name:      req.Name,
			StepsJSON: req.StepsJSON,
		}
		if err := a.funnelStore.Save(item); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(item)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleDeleteSavedFunnel removes a saved funnel by ID.
func (a *APIHandler) HandleDeleteSavedFunnel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	if id == "" {
		http.Error(w, "Missing funnel ID", http.StatusBadRequest)
		return
	}

	if err := a.funnelStore.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// HandleWebVitals returns Core Web Vitals (LCP, INP, CLS p75) breakdown.
func (a *APIHandler) HandleWebVitals(w http.ResponseWriter, r *http.Request) {
	siteID := r.URL.Query().Get("site_id")
	if siteID == "" {
		siteID = "default"
	}
	timeRange := r.URL.Query().Get("range")
	if timeRange == "" {
		timeRange = "7d"
	}

	now := time.Now().UTC()
	start := now.Add(-7 * 24 * time.Hour)
	if timeRange == "24h" {
		start = now.Add(-24 * time.Hour)
	} else if timeRange == "30d" {
		start = now.Add(-30 * 24 * time.Hour)
	}

	report, err := a.eventStore.GetWebVitals(siteID, start, now)
	if err != nil {
		http.Error(w, "Failed to get web vitals: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(report)
}

// HandleEventSearch provides event searching across names and properties.
func (a *APIHandler) HandleEventSearch(w http.ResponseWriter, r *http.Request) {
	siteID := r.URL.Query().Get("site_id")
	if siteID == "" {
		siteID = "default"
	}
	queryStr := r.URL.Query().Get("query")
	eventName := r.URL.Query().Get("event_name")

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	events, total, err := a.eventStore.SearchEvents(siteID, queryStr, eventName, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// HandleSessions returns a paginated list of user sessions with optional rage click filter.
func (a *APIHandler) HandleSessions(w http.ResponseWriter, r *http.Request) {
	siteID := r.URL.Query().Get("site_id")
	if siteID == "" {
		siteID = "default"
	}

	rageOnly := r.URL.Query().Get("rage_only") == "true"

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	sessions, total, err := a.eventStore.GetSessions(siteID, limit, offset, rageOnly)
	if err != nil {
		http.Error(w, "Failed to fetch sessions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// HandleSessionReplay returns decompressed rrweb mutation events for playback.
func (a *APIHandler) HandleSessionReplay(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		sessionID = r.URL.Query().Get("session_id")
	}
	if sessionID == "" {
		http.Error(w, "Missing session ID", http.StatusBadRequest)
		return
	}

	events, err := a.replayStore.GetReplayEvents(sessionID)
	if err != nil {
		http.Error(w, "Failed to retrieve replay: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(events) == 0 {
		http.Error(w, "No replay found for session", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id": sessionID,
		"events":     events,
	})
}

// HandleExportEvents exports events as CSV or JSON.
func (a *APIHandler) HandleExportEvents(w http.ResponseWriter, r *http.Request) {
	siteID := r.URL.Query().Get("site_id")
	if siteID == "" {
		siteID = "default"
	}
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}

	now := time.Now().UTC()
	start := now.Add(-30 * 24 * time.Hour)
	if rng := r.URL.Query().Get("range"); rng == "7d" {
		start = now.Add(-7 * 24 * time.Hour)
	} else if rng == "90d" {
		start = now.Add(-90 * 24 * time.Hour)
	}

	events, err := a.eventStore.ExportEvents(siteID, start, now)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if format == "json" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"microhog-events-%s.json\"", siteID))
		_ = json.NewEncoder(w).Encode(events)
		return
	}

	// CSV Export
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"microhog-events-%s.csv\"", siteID))

	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"id", "site_id", "session_id", "visitor_id", "event_name", "path", "referrer", "browser", "os", "device", "country", "created_at"})

	for _, e := range events {
		_ = writer.Write([]string{
			strconv.FormatInt(e.ID, 10),
			e.SiteID,
			e.SessionID,
			e.VisitorID,
			e.EventName,
			e.Path,
			e.Referrer,
			e.Browser,
			e.OS,
			e.Device,
			e.Country,
			e.CreatedAt.Format(time.RFC3339),
		})
	}
	writer.Flush()
}

// HandleExportSessions exports sessions as CSV or JSON.
func (a *APIHandler) HandleExportSessions(w http.ResponseWriter, r *http.Request) {
	siteID := r.URL.Query().Get("site_id")
	if siteID == "" {
		siteID = "default"
	}
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}

	now := time.Now().UTC()
	start := now.Add(-30 * 24 * time.Hour)
	if rng := r.URL.Query().Get("range"); rng == "7d" {
		start = now.Add(-7 * 24 * time.Hour)
	} else if rng == "90d" {
		start = now.Add(-90 * 24 * time.Hour)
	}

	sessions, err := a.eventStore.ExportSessions(siteID, start, now)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if format == "json" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"microhog-sessions-%s.json\"", siteID))
		_ = json.NewEncoder(w).Encode(sessions)
		return
	}

	// CSV Export
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"microhog-sessions-%s.csv\"", siteID))

	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"id", "site_id", "visitor_id", "started_at", "duration_seconds", "event_count", "entry_path", "exit_path", "browser", "os", "device", "country", "has_replay", "has_rage_click"})

	for _, s := range sessions {
		_ = writer.Write([]string{
			s.ID,
			s.SiteID,
			s.VisitorID,
			s.StartedAt.Format(time.RFC3339),
			strconv.Itoa(s.DurationSeconds),
			strconv.Itoa(s.EventCount),
			s.EntryPath,
			s.ExitPath,
			s.Browser,
			s.OS,
			s.Device,
			s.Country,
			strconv.FormatBool(s.HasReplay),
			strconv.FormatBool(s.HasRageClick),
		})
	}
	writer.Flush()
}

// HandleWebhooks manages registration and listing of alert webhooks.
func (a *APIHandler) HandleWebhooks(w http.ResponseWriter, r *http.Request) {
	siteID := r.URL.Query().Get("site_id")
	if siteID == "" {
		siteID = "default"
	}

	switch r.Method {
	case http.MethodGet:
		list, err := a.webhookStore.ListBySite(siteID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(list)

	case http.MethodPost:
		var req struct {
			ID          string `json:"id"`
			SiteID      string `json:"site_id"`
			Name        string `json:"name"`
			URL         string `json:"url"`
			EventFilter string `json:"event_filter"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}
		if req.SiteID == "" {
			req.SiteID = siteID
		}
		if req.Name == "" {
			req.Name = "Webhook"
		}
		if req.EventFilter == "" {
			req.EventFilter = "all"
		}

		item := &storage.Webhook{
			ID:          req.ID,
			SiteID:      req.SiteID,
			Name:        req.Name,
			URL:         req.URL,
			EventFilter: req.EventFilter,
		}
		if err := a.webhookStore.Save(item); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(item)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleDeleteWebhook removes a registered webhook.
func (a *APIHandler) HandleDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	if id == "" {
		http.Error(w, "Missing webhook ID", http.StatusBadRequest)
		return
	}

	if err := a.webhookStore.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// HandleTestWebhook triggers a test payload to the provided webhook URL.
func (a *APIHandler) HandleTestWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "Valid URL required", http.StatusBadRequest)
		return
	}

	if err := a.webhookDispatcher.Test(req.URL); err != nil {
		http.Error(w, "Webhook test failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Test notification delivered successfully",
	})
}

// HandleSeed triggers synthetic data population.
func (a *APIHandler) HandleSeed(w http.ResponseWriter, r *http.Request) {
	siteID := r.URL.Query().Get("site_id")
	if siteID == "" {
		siteID = "default"
	}

	if err := a.seeder.SeedData(siteID); err != nil {
		http.Error(w, "Failed to seed demo data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Synthetic demo data generated successfully",
	})
}

// HandleSites manages site creation and listing.
func (a *APIHandler) HandleSites(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := a.db.Query("SELECT id, name, domain, created_at FROM sites ORDER BY created_at ASC")
		if err != nil {
			http.Error(w, "Failed to query sites: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type SiteItem struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Domain    string `json:"domain"`
			CreatedAt string `json:"created_at"`
		}

		var sites []SiteItem
		for rows.Next() {
			var s SiteItem
			if err := rows.Scan(&s.ID, &s.Name, &s.Domain, &s.CreatedAt); err != nil {
				continue
			}
			sites = append(sites, s)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sites)

	case http.MethodPost:
		var req struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Domain string `json:"domain"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if req.ID == "" || req.Name == "" {
			http.Error(w, "ID and Name are required", http.StatusBadRequest)
			return
		}

		req.ID = strings.ToLower(strings.TrimSpace(req.ID))
		query := "INSERT INTO sites (id, name, domain) VALUES (?, ?, ?) ON CONFLICT(id) DO UPDATE SET name = excluded.name, domain = excluded.domain"
		if _, err := a.db.Exec(query, req.ID, req.Name, req.Domain); err != nil {
			http.Error(w, "Failed to create site: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(req)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleDeleteSite deletes a site by ID.
func (a *APIHandler) HandleDeleteSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	if id == "" || id == "default" {
		http.Error(w, "Cannot delete default site", http.StatusBadRequest)
		return
	}

	_, err := a.db.Exec("DELETE FROM sites WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
