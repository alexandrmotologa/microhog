package analytics

import (
	"testing"
	"time"

	"github.com/alexandrmotologa/microhog/internal/storage"
)

func TestFunnelEngine(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	es := storage.NewEventStore(db)
	fe := NewFunnelEngine(db)

	now := time.Now().UTC()

	// 3 visitors:
	// Visitor 1 completes all 3 steps: / -> /pricing -> click_signup
	// Visitor 2 completes step 1 and 2: / -> /pricing
	// Visitor 3 completes step 1 only: /
	events := []*storage.Event{
		// Visitor 1
		{
			SiteID:    "default",
			SessionID: "s1",
			VisitorID: "v1",
			EventName: "$pageview",
			Path:      "/",
			CreatedAt: now.Add(-30 * time.Minute),
		},
		{
			SiteID:    "default",
			SessionID: "s1",
			VisitorID: "v1",
			EventName: "$pageview",
			Path:      "/pricing",
			CreatedAt: now.Add(-25 * time.Minute),
		},
		{
			SiteID:    "default",
			SessionID: "s1",
			VisitorID: "v1",
			EventName: "click_signup",
			CreatedAt: now.Add(-20 * time.Minute),
		},

		// Visitor 2
		{
			SiteID:    "default",
			SessionID: "s2",
			VisitorID: "v2",
			EventName: "$pageview",
			Path:      "/",
			CreatedAt: now.Add(-40 * time.Minute),
		},
		{
			SiteID:    "default",
			SessionID: "s2",
			VisitorID: "v2",
			EventName: "$pageview",
			Path:      "/pricing",
			CreatedAt: now.Add(-35 * time.Minute),
		},

		// Visitor 3
		{
			SiteID:    "default",
			SessionID: "s3",
			VisitorID: "v3",
			EventName: "$pageview",
			Path:      "/",
			CreatedAt: now.Add(-50 * time.Minute),
		},
	}

	if err := es.InsertBatch(events); err != nil {
		t.Fatalf("InsertBatch failed: %v", err)
	}

	steps := []FunnelStepDef{
		{Type: "pageview", Path: "/"},
		{Type: "pageview", Path: "/pricing"},
		{Type: "event", Name: "click_signup"},
	}

	res, err := fe.Calculate("default", steps, now.Add(-2*time.Hour), now.Add(2*time.Hour))
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if res.TotalStarted != 3 {
		t.Errorf("Expected 3 started, got %d", res.TotalStarted)
	}
	if res.TotalCompleted != 1 {
		t.Errorf("Expected 1 completed, got %d", res.TotalCompleted)
	}
	if len(res.Steps) != 3 {
		t.Fatalf("Expected 3 steps in result, got %d", len(res.Steps))
	}

	// Step 0: 3 visitors, 100%
	if res.Steps[0].Visitors != 3 || res.Steps[0].ConversionRate != 100.0 {
		t.Errorf("Step 0 mismatch: %+v", res.Steps[0])
	}

	// Step 1: 2 visitors, ~66.7% conversion, ~33.3% dropoff
	if res.Steps[1].Visitors != 2 {
		t.Errorf("Step 1 expected 2 visitors, got %d", res.Steps[1].Visitors)
	}

	// Step 2: 1 visitor, 50% conversion from step 1
	if res.Steps[2].Visitors != 1 || res.Steps[2].ConversionRate != 50.0 {
		t.Errorf("Step 2 mismatch: %+v", res.Steps[2])
	}
}
