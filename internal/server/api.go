package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/alexandrmotologa/microhog/internal/analytics"
	"github.com/alexandrmotologa/microhog/internal/storage"
)

// APIHandler bundles handlers for dashboard REST endpoints.
type APIHandler struct {
	db              *storage.DB
	eventStore      *storage.EventStore
	replayStore     *storage.ReplayStore
	metricsService  *analytics.MetricsService
	funnelEngine    *analytics.FunnelEngine
	realtimeTracker *analytics.RealtimeTracker
	seeder          *analytics.Seeder
	startTime       time.Time
}

// NewAPIHandler creates an APIHandler.
func NewAPIHandler(
	db *storage.DB,
	es *storage.EventStore,
	rs *storage.ReplayStore,
	ms *analytics.MetricsService,
	fe *analytics.FunnelEngine,
	rt *analytics.RealtimeTracker,
	seeder *analytics.Seeder,
) *APIHandler {
	return &APIHandler{
		db:              db,
		eventStore:      es,
		replayStore:     rs,
		metricsService:  ms,
		funnelEngine:    fe,
		realtimeTracker: rt,
		seeder:          seeder,
		startTime:       time.Now(),
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

// HandleSessions returns a paginated list of user sessions.
func (a *APIHandler) HandleSessions(w http.ResponseWriter, r *http.Request) {
	siteID := r.URL.Query().Get("site_id")
	if siteID == "" {
		siteID = "default"
	}

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

	sessions, total, err := a.eventStore.GetSessions(siteID, limit, offset)
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

// HandleSites returns the list of configured sites.
func (a *APIHandler) HandleSites(w http.ResponseWriter, r *http.Request) {
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
}
