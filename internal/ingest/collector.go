package ingest

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alexandrmotologa/microhog/internal/analytics"
	"github.com/alexandrmotologa/microhog/internal/storage"
)

// RawEventPayload represents client-submitted event JSON.
type RawEventPayload struct {
	SiteID     string                 `json:"site_id"`
	SessionID  string                 `json:"session_id"`
	Event      string                 `json:"event"`
	URL        string                 `json:"url"`
	Path       string                 `json:"path"`
	Referrer   string                 `json:"referrer"`
	Properties map[string]interface{} `json:"properties"`
}

// ReplayChunkPayload represents incoming rrweb DOM session chunks.
type ReplayChunkPayload struct {
	SessionID  string          `json:"session_id"`
	ChunkIndex int             `json:"chunk_index"`
	EventCount int             `json:"event_count"`
	Data       json.RawMessage `json:"data"`
	Events     json.RawMessage `json:"events"`
}

// Collector exposes HTTP endpoints for receiving event and replay data.
type Collector struct {
	anonymizer      *Anonymizer
	bufferPool      *BufferPool
	replayStore     *storage.ReplayStore
	realtimeTracker *analytics.RealtimeTracker
}

// NewCollector constructs a Collector instance.
func NewCollector(
	anon *Anonymizer,
	bp *BufferPool,
	rs *storage.ReplayStore,
	rt *analytics.RealtimeTracker,
) *Collector {
	return &Collector{
		anonymizer:      anon,
		bufferPool:      bp,
		replayStore:     rs,
		realtimeTracker: rt,
	}
}

// HandleEvent processes incoming single or batch event payloads from client trackers.
func (c *Collector) HandleEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1024*1024)) // 1MB limit
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	var rawEvents []RawEventPayload

	// Check if array or single object
	trimmed := strings.TrimSpace(string(body))
	if strings.HasPrefix(trimmed, "[") {
		if err := json.Unmarshal(body, &rawEvents); err != nil {
			http.Error(w, "Invalid JSON array", http.StatusBadRequest)
			return
		}
	} else {
		var single RawEventPayload
		if err := json.Unmarshal(body, &single); err != nil {
			http.Error(w, "Invalid JSON object", http.StatusBadRequest)
			return
		}
		rawEvents = append(rawEvents, single)
	}

	clientInfo := c.anonymizer.ParseRequest(r, "")
	received := 0

	for _, raw := range rawEvents {
		if raw.Event == "" {
			continue
		}
		siteID := raw.SiteID
		if siteID == "" {
			siteID = "default"
		}

		sessionID := raw.SessionID
		if sessionID == "" {
			sessionID = c.anonymizer.HashVisitorID(clientInfo.IP, clientInfo.UserAgent, siteID)[:16]
		}

		visitorID := c.anonymizer.HashVisitorID(clientInfo.IP, clientInfo.UserAgent, siteID)

		ref := raw.Referrer
		refDomain := clientInfo.ReferrerDomain
		if ref != "" {
			refDomain = extractDomain(ref)
		}

		path := raw.Path
		if path == "" && raw.URL != "" {
			path = extractPath(raw.URL)
		}
		if path == "" {
			path = "/"
		}

		propsJSON := "{}"
		if len(raw.Properties) > 0 {
			if b, err := json.Marshal(raw.Properties); err == nil {
				propsJSON = string(b)
			}
		}

		event := &storage.Event{
			SiteID:         siteID,
			SessionID:      sessionID,
			VisitorID:      visitorID,
			EventName:      raw.Event,
			URL:            raw.URL,
			Path:           path,
			Referrer:       ref,
			ReferrerDomain: refDomain,
			Browser:        clientInfo.Browser,
			OS:             clientInfo.OS,
			Device:         clientInfo.Device,
			Country:        clientInfo.Country,
			PropertiesJSON: propsJSON,
			CreatedAt:      time.Now().UTC(),
		}

		// Enqueue for batch write
		c.bufferPool.Enqueue(event)

		// Update real-time tracker
		if c.realtimeTracker != nil {
			c.realtimeTracker.RecordVisit(event)
		}

		received++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "ok",
		"received": received,
	})
}

// HandleReplayChunk receives compressed or raw DOM session replay records from recorder.ts.
func (c *Collector) HandleReplayChunk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 10*1024*1024)) // 10MB limit for replays
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload ReplayChunkPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if payload.SessionID == "" {
		http.Error(w, "Missing session_id", http.StatusBadRequest)
		return
	}

	var rawEventsJSON []byte
	if len(payload.Events) > 0 {
		rawEventsJSON = payload.Events
	} else if len(payload.Data) > 0 {
		rawEventsJSON = payload.Data
	} else {
		http.Error(w, "Empty replay events data", http.StatusBadRequest)
		return
	}

	count := payload.EventCount
	if count <= 0 {
		var list []interface{}
		if err := json.Unmarshal(rawEventsJSON, &list); err == nil {
			count = len(list)
		}
	}

	if err := c.replayStore.SaveChunk(payload.SessionID, payload.ChunkIndex, rawEventsJSON, count); err != nil {
		http.Error(w, "Failed to save replay chunk: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func extractPath(rawURL string) string {
	idx := strings.Index(rawURL, "://")
	rest := rawURL
	if idx != -1 {
		rest = rawURL[idx+3:]
	}
	slashIdx := strings.Index(rest, "/")
	if slashIdx == -1 {
		return "/"
	}
	path := rest[slashIdx:]
	if qIdx := strings.Index(path, "?"); qIdx != -1 {
		path = path[:qIdx]
	}
	if hIdx := strings.Index(path, "#"); hIdx != -1 {
		path = path[:hIdx]
	}
	return path
}
