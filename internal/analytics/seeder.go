package analytics

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/alexandrmotologa/microhog/internal/storage"
)

// Seeder inserts realistic synthetic analytics data for testing and demonstration.
type Seeder struct {
	db          *storage.DB
	eventStore  *storage.EventStore
	replayStore *storage.ReplayStore
}

// NewSeeder creates a new Seeder.
func NewSeeder(db *storage.DB, es *storage.EventStore, rs *storage.ReplayStore) *Seeder {
	return &Seeder{
		db:          db,
		eventStore:  es,
		replayStore: rs,
	}
}

// SeedData populates the database with realistic traffic across the past 7 days.
func (s *Seeder) SeedData(siteID string) error {
	if siteID == "" {
		siteID = "default"
	}

	r := rand.New(rand.NewSource(42)) // Deterministic seed for reproducible data

	paths := []string{
		"/",
		"/pricing",
		"/features",
		"/docs",
		"/blog/micro-saas-analytics",
		"/signup",
		"/app/dashboard",
	}

	referrers := []struct {
		url    string
		domain string
		weight int
	}{
		{"https://news.ycombinator.com", "news.ycombinator.com", 25},
		{"https://www.google.com/search?q=lightweight+analytics", "google.com", 30},
		{"https://github.com/alexandrmotologa/microhog", "github.com", 20},
		{"https://x.com/mtlglabs", "x.com", 15},
		{"", "direct", 40},
	}

	browsers := []string{"Chrome", "Safari", "Firefox", "Edge"}
	osList := []string{"macOS", "Windows", "Linux", "iOS", "Android"}
	countries := []string{"US", "DE", "GB", "FR", "RO", "NL", "CA", "JP"}

	now := time.Now().UTC()
	var allEvents []*storage.Event

	totalSessions := 240
	var sampleSessionID string

	for i := 0; i < totalSessions; i++ {
		// Session spread across the last 7 days
		hoursAgo := r.Float64() * 168.0 // 7 days = 168h
		sessionStart := now.Add(-time.Duration(hoursAgo * float64(time.Hour)))

		sessionID := fmt.Sprintf("sess_%04d_%x", i, sessionStart.Unix())
		visitorID := fmt.Sprintf("vis_%04d_%x", i%95, (i%95)*1337)

		// Choose metadata
		browser := browsers[r.Intn(len(browsers))]
		osName := osList[r.Intn(len(osList))]
		device := "desktop"
		if osName == "iOS" || osName == "Android" {
			device = "mobile"
		}
		country := countries[r.Intn(len(countries))]

		refChoice := referrers[r.Intn(len(referrers))]

		// Pick user journey
		// Funnel path: / -> /pricing -> click_signup -> /signup -> /app/dashboard
		var journey []struct {
			name      string
			path      string
			delaySecs int
		}

		journey = append(journey, struct {
			name      string
			path      string
			delaySecs int
		}{"$pageview", "/", 0})

		// 60% chance to view pricing
		if r.Float64() < 0.60 {
			journey = append(journey, struct {
				name      string
				path      string
				delaySecs int
			}{"$pageview", "/pricing", 15 + r.Intn(45)})

			// 35% of those click signup
			if r.Float64() < 0.35 {
				journey = append(journey, struct {
					name      string
					path      string
					delaySecs int
				}{"click_signup", "/pricing", 5 + r.Intn(20)})

				journey = append(journey, struct {
					name      string
					path      string
					delaySecs int
				}{"$pageview", "/signup", 3 + r.Intn(10)})

				// 50% finish signup and reach app
				if r.Float64() < 0.50 {
					journey = append(journey, struct {
						name      string
						path      string
						delaySecs int
					}{"signup_completed", "/signup", 10 + r.Intn(30)})

					journey = append(journey, struct {
						name      string
						path      string
						delaySecs int
					}{"$pageview", "/app/dashboard", 2 + r.Intn(5)})
				}
			}
		} else {
			// Some casual browsing
			extraPages := 1 + r.Intn(3)
			elapsed := 10
			for p := 0; p < extraPages; p++ {
				page := paths[r.Intn(len(paths))]
				elapsed += 10 + r.Intn(30)
				journey = append(journey, struct {
					name      string
					path      string
					delaySecs int
				}{"$pageview", page, elapsed})
			}
		}

		var currentSessionTime = sessionStart
		for _, step := range journey {
			currentSessionTime = currentSessionTime.Add(time.Duration(step.delaySecs) * time.Second)
			allEvents = append(allEvents, &storage.Event{
				SiteID:         siteID,
				SessionID:      sessionID,
				VisitorID:      visitorID,
				EventName:      step.name,
				URL:            "https://example.com" + step.path,
				Path:           step.path,
				Referrer:       refChoice.url,
				ReferrerDomain: refChoice.domain,
				Browser:        browser,
				OS:             osName,
				Device:         device,
				Country:        country,
				PropertiesJSON: "{}",
				CreatedAt:      currentSessionTime,
			})
		}

		if i == 0 {
			sampleSessionID = sessionID
		}
	}

	// Insert in batches of 100
	for i := 0; i < len(allEvents); i += 100 {
		end := i + 100
		if end > len(allEvents) {
			end = len(allEvents)
		}
		if err := s.eventStore.InsertBatch(allEvents[i:end]); err != nil {
			return fmt.Errorf("failed to insert seed batch: %w", err)
		}
	}

	// Create sample rrweb session replay for the first session
	if sampleSessionID != "" {
		if err := s.seedSampleReplay(sampleSessionID); err != nil {
			return fmt.Errorf("failed to seed sample replay: %w", err)
		}
	}

	return nil
}

func (s *Seeder) seedSampleReplay(sessionID string) error {
	// Synthesize a valid rrweb event stream representing a real browser DOM recording
	// Event types in rrweb:
	// 0: DomContentLoaded, 1: Load, 2: FullSnapshot, 3: IncrementalSnapshot (mouse move, mouse click, scroll)
	baseTime := time.Now().UnixMilli()

	rrwebEvents := []map[string]interface{}{
		{
			"type":      0,
			"timestamp": baseTime,
			"data":      map[string]interface{}{},
		},
		{
			"type":      1,
			"timestamp": baseTime + 100,
			"data":      map[string]interface{}{},
		},
		{
			"type":      2, // FullSnapshot of document
			"timestamp": baseTime + 200,
			"data": map[string]interface{}{
				"node": map[string]interface{}{
					"id":   1,
					"type": 0, // Document
					"childNodes": []map[string]interface{}{
						{
							"id":        2,
							"type":      1, // DocumentType
							"name":      "html",
							"publicId":  "",
							"systemId":  "",
						},
						{
							"id":       3,
							"type":     2, // Element html
							"tagName":  "html",
							"attributes": map[string]interface{}{"lang": "en"},
							"childNodes": []map[string]interface{}{
								{
									"id":      4,
									"type":    2, // Element head
									"tagName": "head",
									"childNodes": []map[string]interface{}{
										{
											"id":       5,
											"type":     2, // style
											"tagName":  "style",
											"childNodes": []map[string]interface{}{
												{
													"id":          6,
													"type":        3, // Text
													"textContent": "body { font-family: -apple-system, sans-serif; background: #0f172a; color: #f8fafc; padding: 40px; margin: 0; } .hero { max-width: 600px; margin: 0 auto; text-align: center; } .btn { display: inline-block; background: #3b82f6; color: white; padding: 12px 24px; border-radius: 8px; text-decoration: none; font-weight: 600; cursor: pointer; border: none; margin-top: 24px; transition: transform 0.2s; } .btn:hover { background: #2563eb; transform: scale(1.05); } .card { background: #1e293b; padding: 24px; border-radius: 12px; margin-top: 32px; border: 1px solid #334155; } input { width: 100%; box-sizing: border-box; background: #0f172a; border: 1px solid #475569; color: white; padding: 10px; border-radius: 6px; margin-top: 8px; }",
												},
											},
										},
									},
								},
								{
									"id":      7,
									"type":    2, // Element body
									"tagName": "body",
									"childNodes": []map[string]interface{}{
										{
											"id":       8,
											"type":     2,
											"tagName":  "div",
											"attributes": map[string]interface{}{"class": "hero"},
											"childNodes": []map[string]interface{}{
												{
													"id":       9,
													"type":     2,
													"tagName":  "h1",
													"childNodes": []map[string]interface{}{
														{
															"id":          10,
															"type":        3,
															"textContent": "MicroHog Replay Preview",
														},
													},
												},
												{
													"id":       11,
													"type":     2,
													"tagName":  "p",
													"childNodes": []map[string]interface{}{
														{
															"id":          12,
															"type":        3,
															"textContent": "Watching a recorded user interaction session captured with rrweb.",
														},
													},
												},
												{
													"id":       13,
													"type":     2,
													"tagName":  "div",
													"attributes": map[string]interface{}{"class": "card"},
													"childNodes": []map[string]interface{}{
														{
															"id":       14,
															"type":     2,
															"tagName":  "label",
															"childNodes": []map[string]interface{}{
																{"id": 15, "type": 3, "textContent": "Workspace Name"},
															},
														},
														{
															"id":       16,
															"type":     2,
															"tagName":  "input",
															"attributes": map[string]interface{}{
																"type":        "text",
																"value":       "Acme SaaS",
																"placeholder": "Enter name",
															},
														},
														{
															"id":       17,
															"type":     2,
															"tagName":  "button",
															"attributes": map[string]interface{}{
																"class": "btn",
																"id":    "start-btn",
															},
															"childNodes": []map[string]interface{}{
																{"id": 18, "type": 3, "textContent": "Deploy Instance"},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"initialOffset": map[string]interface{}{"top": 0, "left": 0},
			},
		},
	}

	// Add mouse movement and click snapshots
	// Mouse movement simulation
	coords := [][2]int{
		{200, 150}, {250, 180}, {300, 220}, {350, 270}, {400, 310},
		{420, 350}, {430, 380}, {435, 410}, {440, 420},
	}

	t := baseTime + 400
	for _, pt := range coords {
		t += 150
		rrwebEvents = append(rrwebEvents, map[string]interface{}{
			"type":      3, // IncrementalSnapshot
			"timestamp": t,
			"data": map[string]interface{}{
				"source": 1, // MouseMove
				"positions": []map[string]interface{}{
					{"x": pt[0], "y": pt[1], "id": 7, "timeOffset": 0},
				},
			},
		})
	}

	// Mouse click on button at t + 300
	t += 300
	rrwebEvents = append(rrwebEvents, map[string]interface{}{
		"type":      3,
		"timestamp": t,
		"data": map[string]interface{}{
			"source": 2, // MouseInteraction
			"type":   2, // Click
			"id":     17,
			"x":      440,
			"y":      420,
		},
	})

	rawJSON, err := json.Marshal(rrwebEvents)
	if err != nil {
		return err
	}

	return s.replayStore.SaveChunk(sessionID, 0, rawJSON, len(rrwebEvents))
}
