package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexandrmotologa/microhog/internal/analytics"
	"github.com/alexandrmotologa/microhog/internal/ingest"
	"github.com/alexandrmotologa/microhog/internal/storage"
)

func TestFullHTTPFlow(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
	}
	defer db.Close()

	es := storage.NewEventStore(db)
	rs, _ := storage.NewReplayStore(db, es)
	ms := analytics.NewMetricsService(es)
	fe := analytics.NewFunnelEngine(db)
	rt := analytics.NewRealtimeTracker()
	seeder := analytics.NewSeeder(db, es, rs)

	anon := ingest.NewAnonymizer("test_secret")
	bp := ingest.NewBufferPool(es, 100, 10, 50*time.Millisecond)
	defer bp.Stop()

	col := ingest.NewCollector(anon, bp, rs, rt)
	api := NewAPIHandler(db, es, rs, ms, fe, rt, seeder)

	staticFS := GetEmbeddedFS()
	router := NewRouter(&RouterConfig{
		Collector:       col,
		APIHandler:      api,
		RealtimeTracker: rt,
		StaticFileFS:    staticFS,
	})

	server := httptest.NewServer(router)
	defer server.Close()

	// 1. Test /health
	res, err := http.Get(server.URL + "/health")
	if err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("Health check failed: %v", err)
	}

	// 2. Test POST /e (single event)
	eventPayload := map[string]interface{}{
		"site_id":    "default",
		"session_id": "s_http_1",
		"event":      "$pageview",
		"url":        "http://example.com/test",
		"path":       "/test",
	}
	body, _ := json.Marshal(eventPayload)

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/e", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err = http.DefaultClient.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("POST /e failed: %v", err)
	}

	// Wait for buffer pool flush
	time.Sleep(100 * time.Millisecond)

	// 3. Test GET /api/realtime
	res, err = http.Get(server.URL + "/api/realtime?site_id=default")
	if err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/realtime failed: %v", err)
	}

	// 4. Test GET /api/metrics
	res, err = http.Get(server.URL + "/api/metrics?site_id=default&range=24h")
	if err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/metrics failed: %v", err)
	}
	var metricsRes analytics.MetricsPayload
	if err := json.NewDecoder(res.Body).Decode(&metricsRes); err != nil {
		t.Fatalf("Failed to decode metrics response: %v", err)
	}
	if metricsRes.Summary.Pageviews != 1 {
		t.Errorf("Expected 1 pageview, got %d", metricsRes.Summary.Pageviews)
	}

	// 5. Test POST /api/replay/chunk
	replayPayload := map[string]interface{}{
		"session_id":  "s_http_1",
		"chunk_index": 0,
		"events": []map[string]interface{}{
			{"type": 0, "timestamp": 12345},
			{"type": 2, "timestamp": 12350},
		},
	}
	replayBody, _ := json.Marshal(replayPayload)
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/api/replay/chunk", bytes.NewReader(replayBody))
	req.Header.Set("Content-Type", "application/json")
	res, err = http.DefaultClient.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("POST /api/replay/chunk failed: %v", err)
	}

	// 6. Test GET /api/sessions/s_http_1/replay
	res, err = http.Get(server.URL + "/api/sessions/s_http_1/replay")
	if err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/sessions/:id/replay failed: %v", err)
	}

	// 7. Test Tracker Script Serving
	res, err = http.Get(server.URL + "/microhog.js")
	if err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("GET /microhog.js failed: %v", err)
	}

	// 8. Test Embedded Web Studio Dashboard Serving
	res, err = http.Get(server.URL + "/")
	if err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("GET / failed: %v", err)
	}
}
