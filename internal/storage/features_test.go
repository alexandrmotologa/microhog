package storage

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFeaturesAndEnhancements(t *testing.T) {
	db, err := OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
	}
	defer db.Close()

	es := NewEventStore(db)
	fs := NewFunnelStore(db)
	ws := NewWebhookStore(db)

	now := time.Now().UTC()

	// 1. Test Rage Click recording & filtering
	events := []*Event{
		{
			SiteID:         "default",
			SessionID:      "sess_normal",
			VisitorID:      "vis_1",
			EventName:      "$pageview",
			Path:           "/",
			PropertiesJSON: "{}",
			CreatedAt:      now.Add(-20 * time.Minute),
		},
		{
			SiteID:         "default",
			SessionID:      "sess_rage",
			VisitorID:      "vis_2",
			EventName:      "$rage_click",
			Path:           "/pricing",
			PropertiesJSON: `{"target_tag":"button","clicks":4,"rage_click":true}`,
			CreatedAt:      now.Add(-10 * time.Minute),
		},
	}
	if err := es.InsertBatch(events); err != nil {
		t.Fatalf("InsertBatch failed: %v", err)
	}

	// Verify all sessions
	_, totalAll, err := es.GetSessions("default", 10, 0, false)
	if err != nil || totalAll != 2 {
		t.Fatalf("Expected 2 total sessions, got %d (err: %v)", totalAll, err)
	}

	// Verify rage-only sessions filter
	rageSessions, totalRage, err := es.GetSessions("default", 10, 0, true)
	if err != nil || totalRage != 1 {
		t.Fatalf("Expected 1 rage session, got %d", totalRage)
	}
	if !rageSessions[0].HasRageClick {
		t.Errorf("Expected rage session to have has_rage_click=true")
	}

	// 2. Test Core Web Vitals aggregation
	vitalsEvents := []*Event{
		{
			SiteID:         "default",
			SessionID:      "sess_1",
			VisitorID:      "vis_1",
			EventName:      "$web_vitals",
			Path:           "/",
			PropertiesJSON: `{"lcp":1500.0,"inp":80.0,"cls":0.02}`,
			CreatedAt:      now.Add(-5 * time.Minute),
		},
		{
			SiteID:         "default",
			SessionID:      "sess_2",
			VisitorID:      "vis_2",
			EventName:      "$web_vitals",
			Path:           "/",
			PropertiesJSON: `{"lcp":2100.0,"inp":120.0,"cls":0.04}`,
			CreatedAt:      now.Add(-3 * time.Minute),
		},
	}
	if err := es.InsertBatch(vitalsEvents); err != nil {
		t.Fatalf("InsertBatch vitals failed: %v", err)
	}

	vitalsReport, err := es.GetWebVitals("default", now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetWebVitals failed: %v", err)
	}
	if vitalsReport.LCP.Count != 2 {
		t.Errorf("Expected 2 LCP samples, got %d", vitalsReport.LCP.Count)
	}
	if vitalsReport.LCP.Rating != "good" {
		t.Errorf("Expected good rating for LCP, got %s", vitalsReport.LCP.Rating)
	}
	if len(vitalsReport.Pages) != 1 || vitalsReport.Pages[0].Path != "/" {
		t.Errorf("Expected 1 page report for '/', got %+v", vitalsReport.Pages)
	}

	// 3. Test Saved Funnels Persistence
	steps := []map[string]string{
		{"type": "pageview", "path": "/"},
		{"type": "pageview", "path": "/signup"},
	}
	stepsJSON, _ := json.Marshal(steps)
	savedF := &SavedFunnel{
		SiteID:    "default",
		Name:      "Signup Funnel",
		StepsJSON: string(stepsJSON),
	}
	if err := fs.Save(savedF); err != nil {
		t.Fatalf("SaveFunnel failed: %v", err)
	}

	funnelsList, err := fs.ListBySite("default")
	if err != nil || len(funnelsList) != 1 {
		t.Fatalf("Expected 1 saved funnel, got %d", len(funnelsList))
	}
	if funnelsList[0].Name != "Signup Funnel" {
		t.Errorf("Funnel name mismatch: %s", funnelsList[0].Name)
	}

	if err := fs.Delete(funnelsList[0].ID); err != nil {
		t.Fatalf("Delete funnel failed: %v", err)
	}
	remainingFunnels, _ := fs.ListBySite("default")
	if len(remainingFunnels) != 0 {
		t.Errorf("Expected 0 funnels after deletion")
	}

	// 4. Test Webhooks Persistence
	wh := &Webhook{
		SiteID:      "default",
		Name:        "Slack #growth",
		URL:         "https://hooks.slack.com/services/test",
		EventFilter: "signup_completed",
	}
	if err := ws.Save(wh); err != nil {
		t.Fatalf("SaveWebhook failed: %v", err)
	}
	whList, err := ws.ListBySite("default")
	if err != nil || len(whList) != 1 {
		t.Fatalf("Expected 1 webhook, got %d", len(whList))
	}
	if err := ws.Delete(whList[0].ID); err != nil {
		t.Fatalf("Delete webhook failed: %v", err)
	}

	// 5. Test Event Search
	matched, total, err := es.SearchEvents("default", "pricing", "all", 10, 0)
	if err != nil {
		t.Fatalf("SearchEvents failed: %v", err)
	}
	if total != 1 || len(matched) != 1 {
		t.Errorf("Expected 1 matched event for 'pricing', got %d", total)
	}

	// 6. Test CSV Export queries
	expEvents, err := es.ExportEvents("default", now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil || len(expEvents) < 3 {
		t.Errorf("ExportEvents failed: %v (count: %d)", err, len(expEvents))
	}
	expSessions, err := es.ExportSessions("default", now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil || len(expSessions) != 4 {
		t.Errorf("ExportSessions failed: %v (count: %d)", err, len(expSessions))
	}
}
