package storage

import (
	"encoding/json"
	"testing"
	"time"
)

func TestStorageLifecycle(t *testing.T) {
	// 1. Open in-memory or temporary SQLite
	db, err := OpenDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	es := NewEventStore(db)
	rs, err := NewReplayStore(db, es)
	if err != nil {
		t.Fatalf("Failed to create replay store: %v", err)
	}

	// 2. Insert batch of events
	now := time.Now().UTC()
	events := []*Event{
		{
			SiteID:         "default",
			SessionID:      "sess_test_1",
			VisitorID:      "vis_1",
			EventName:      "$pageview",
			Path:           "/",
			Referrer:       "https://google.com",
			ReferrerDomain: "google.com",
			Browser:        "Chrome",
			OS:             "macOS",
			Device:         "desktop",
			Country:        "US",
			CreatedAt:      now.Add(-10 * time.Minute),
		},
		{
			SiteID:         "default",
			SessionID:      "sess_test_1",
			VisitorID:      "vis_1",
			EventName:      "$pageview",
			Path:           "/pricing",
			Referrer:       "",
			ReferrerDomain: "direct",
			Browser:        "Chrome",
			OS:             "macOS",
			Device:         "desktop",
			Country:        "US",
			CreatedAt:      now.Add(-5 * time.Minute),
		},
	}

	if err := es.InsertBatch(events); err != nil {
		t.Fatalf("InsertBatch failed: %v", err)
	}

	// 3. Verify metrics summary
	summary, err := es.GetMetricsSummary("default", now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetMetricsSummary failed: %v", err)
	}
	if summary.Pageviews != 2 {
		t.Errorf("Expected 2 pageviews, got %d", summary.Pageviews)
	}
	if summary.UniqueVisitors != 1 {
		t.Errorf("Expected 1 unique visitor, got %d", summary.UniqueVisitors)
	}

	// 4. Test Top Pages
	pages, err := es.GetTopPages("default", now.Add(-1*time.Hour), now.Add(1*time.Hour), 10)
	if err != nil {
		t.Fatalf("GetTopPages failed: %v", err)
	}
	if len(pages) != 2 {
		t.Errorf("Expected 2 pages, got %d", len(pages))
	}

	// 5. Test Sessions
	sessions, total, err := es.GetSessions("default", 10, 0, false)
	if err != nil {
		t.Fatalf("GetSessions failed: %v", err)
	}
	if total != 1 {
		t.Errorf("Expected 1 session, got %d", total)
	}
	if len(sessions) != 1 || sessions[0].EventCount != 2 {
		t.Errorf("Session event count mismatch: %+v", sessions[0])
	}

	// 6. Test ReplayStore (save and decompress chunks)
	dummyRRWeb := []map[string]interface{}{
		{"type": 0, "timestamp": now.UnixMilli()},
		{"type": 2, "timestamp": now.UnixMilli() + 50},
	}
	rawJSON, _ := json.Marshal(dummyRRWeb)

	if err := rs.SaveChunk("sess_test_1", 0, rawJSON, len(dummyRRWeb)); err != nil {
		t.Fatalf("SaveChunk failed: %v", err)
	}

	retrieved, err := rs.GetReplayEvents("sess_test_1")
	if err != nil {
		t.Fatalf("GetReplayEvents failed: %v", err)
	}
	if len(retrieved) != 2 {
		t.Fatalf("Expected 2 replay frames, got %d", len(retrieved))
	}

	// 7. Verify session has_replay was set to true
	updatedSessions, _, _ := es.GetSessions("default", 1, 0, false)
	if len(updatedSessions) == 0 || !updatedSessions[0].HasReplay {
		t.Errorf("Expected session has_replay to be true")
	}
}
