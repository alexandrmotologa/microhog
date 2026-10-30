package analytics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/alexandrmotologa/microhog/internal/storage"
)

// ActiveSession tracks a currently browsing visitor in memory.
type ActiveSession struct {
	SessionID   string    `json:"session_id"`
	VisitorID   string    `json:"visitor_id"`
	SiteID      string    `json:"site_id"`
	CurrentPath string    `json:"current_path"`
	Country     string    `json:"country"`
	Browser     string    `json:"browser"`
	Device      string    `json:"device"`
	LastSeen    time.Time `json:"last_seen"`
}

// RealtimeStats summarizes current live activity for the dashboard.
type RealtimeStats struct {
	ActiveVisitors int                      `json:"active_visitors"`
	ActivePages    []*storage.PageStat      `json:"active_pages"`
	Countries      []*storage.BreakdownStat `json:"countries"`
	RecentEvents   []*storage.Event         `json:"recent_events"`
}

// RealtimeTracker maintains the 5-minute sliding window of active visitors and broadcasts SSE events.
type RealtimeTracker struct {
	mu           sync.RWMutex
	sessions     map[string]*ActiveSession // key: siteID + ":" + sessionID
	recentEvents []*storage.Event
	clients      map[chan []byte]string // channel -> siteID
	clientsMu    sync.RWMutex
}

// NewRealtimeTracker initializes the in-memory tracker.
func NewRealtimeTracker() *RealtimeTracker {
	rt := &RealtimeTracker{
		sessions:     make(map[string]*ActiveSession),
		recentEvents: make([]*storage.Event, 0, 50),
		clients:      make(map[chan []byte]string),
	}
	go rt.cleanupLoop()
	return rt
}

// RecordVisit updates active session information and forwards event to SSE listeners.
func (rt *RealtimeTracker) RecordVisit(e *storage.Event) {
	key := e.SiteID + ":" + e.SessionID
	now := time.Now().UTC()

	rt.mu.Lock()
	sess, exists := rt.sessions[key]
	if !exists {
		sess = &ActiveSession{
			SessionID: e.SessionID,
			VisitorID: e.VisitorID,
			SiteID:    e.SiteID,
			Country:   e.Country,
			Browser:   e.Browser,
			Device:    e.Device,
		}
		rt.sessions[key] = sess
	}
	sess.LastSeen = now
	if e.Path != "" {
		sess.CurrentPath = e.Path
	}

	// Keep last 30 events for instant live feed
	rt.recentEvents = append([]*storage.Event{e}, rt.recentEvents...)
	if len(rt.recentEvents) > 30 {
		rt.recentEvents = rt.recentEvents[:30]
	}
	rt.mu.Unlock()

	// Broadcast to SSE subscribers
	rt.broadcastEvent(e)
}

func (rt *RealtimeTracker) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().UTC().Add(-5 * time.Minute)
		rt.mu.Lock()
		for k, sess := range rt.sessions {
			if sess.LastSeen.Before(cutoff) {
				delete(rt.sessions, k)
			}
		}
		rt.mu.Unlock()
	}
}

// GetStats compiles current active sessions for a site into structured metrics.
func (rt *RealtimeTracker) GetStats(siteID string) *RealtimeStats {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	cutoff := time.Now().UTC().Add(-5 * time.Minute)
	visitorsMap := make(map[string]struct{})
	pageCounts := make(map[string]int64)
	countryCounts := make(map[string]int64)

	for _, sess := range rt.sessions {
		if sess.SiteID != siteID && siteID != "" && siteID != "all" {
			continue
		}
		if sess.LastSeen.Before(cutoff) {
			continue
		}

		visitorsMap[sess.VisitorID] = struct{}{}
		if sess.CurrentPath != "" {
			pageCounts[sess.CurrentPath]++
		}
		if sess.Country != "" {
			countryCounts[sess.Country]++
		}
	}

	// Pages breakdown
	var activePages []*storage.PageStat
	for path, count := range pageCounts {
		activePages = append(activePages, &storage.PageStat{
			Path:  path,
			Views: count,
		})
	}
	sort.Slice(activePages, func(i, j int) bool {
		return activePages[i].Views > activePages[j].Views
	})
	if len(activePages) > 10 {
		activePages = activePages[:10]
	}

	// Countries breakdown
	var countries []*storage.BreakdownStat
	for code, count := range countryCounts {
		countries = append(countries, &storage.BreakdownStat{
			Name:  code,
			Count: count,
		})
	}
	sort.Slice(countries, func(i, j int) bool {
		return countries[i].Count > countries[j].Count
	})
	if len(countries) > 10 {
		countries = countries[:10]
	}

	// Filter recent events for this site
	var siteRecent []*storage.Event
	for _, e := range rt.recentEvents {
		if sessSite := e.SiteID; sessSite == siteID || siteID == "" || siteID == "all" {
			siteRecent = append(siteRecent, e)
		}
		if len(siteRecent) >= 15 {
			break
		}
	}

	return &RealtimeStats{
		ActiveVisitors: len(visitorsMap),
		ActivePages:    activePages,
		Countries:      countries,
		RecentEvents:   siteRecent,
	}
}

// HandleSSE serves a Server-Sent Events stream to dashboard clients.
func (rt *RealtimeTracker) HandleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	siteID := r.URL.Query().Get("site_id")
	if siteID == "" {
		siteID = "default"
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	eventChan := make(chan []byte, 32)
	rt.clientsMu.Lock()
	rt.clients[eventChan] = siteID
	rt.clientsMu.Unlock()

	defer func() {
		rt.clientsMu.Lock()
		delete(rt.clients, eventChan)
		rt.clientsMu.Unlock()
		close(eventChan)
	}()

	// Send initial snapshot
	initialStats := rt.GetStats(siteID)
	data, _ := json.Marshal(initialStats)
	fmt.Fprintf(w, "event: snapshot\ndata: %s\n\n", data)
	flusher.Flush()

	notify := r.Context().Done()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-notify:
			return
		case msg, ok := <-eventChan:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: live_event\ndata: %s\n\n", msg)
			flusher.Flush()
		case <-heartbeat.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

func (rt *RealtimeTracker) broadcastEvent(e *storage.Event) {
	payload, err := json.Marshal(e)
	if err != nil {
		return
	}

	rt.clientsMu.RLock()
	defer rt.clientsMu.RUnlock()

	for ch, siteID := range rt.clients {
		if siteID == e.SiteID || siteID == "all" || siteID == "" {
			select {
			case ch <- payload:
			default:
				// Avoid blocking if client buffer is slow
			}
		}
	}
}
