package analytics

import (
	"fmt"
	"strings"
	"time"

	"github.com/alexandrmotologa/microhog/internal/storage"
)

// FunnelStepDef defines one sequential step in a conversion funnel.
type FunnelStepDef struct {
	Type string `json:"type"` // "pageview" or "event"
	Path string `json:"path"` // e.g. "/pricing" (for pageviews)
	Name string `json:"name"` // e.g. "signup_completed" (for custom events)
}

// FunnelStepResult holds statistics for one completed funnel stage.
type FunnelStepResult struct {
	StepIndex      int     `json:"step_index"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Visitors       int64   `json:"visitors"`
	ConversionRate float64 `json:"conversion_rate"`
	DropOffRate    float64 `json:"drop_off_rate"`
}

// FunnelResult represents the complete funnel analysis.
type FunnelResult struct {
	Steps                 []*FunnelStepResult `json:"steps"`
	OverallConversionRate float64             `json:"overall_conversion_rate"`
	TotalStarted          int64               `json:"total_started"`
	TotalCompleted        int64               `json:"total_completed"`
}

// FunnelEngine computes conversion rates across ordered events.
type FunnelEngine struct {
	db *storage.DB
}

// NewFunnelEngine constructs a FunnelEngine instance.
func NewFunnelEngine(db *storage.DB) *FunnelEngine {
	return &FunnelEngine{db: db}
}

func parseAnyTime(val interface{}) time.Time {
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
	case []byte:
		return parseAnyTime(string(v))
	}
	return time.Time{}
}

// Calculate computes step-by-step conversion for given step definitions.
func (f *FunnelEngine) Calculate(siteID string, steps []FunnelStepDef, start, end time.Time) (*FunnelResult, error) {
	if len(steps) < 2 {
		return nil, fmt.Errorf("funnel requires at least 2 steps")
	}

	startTime := start.UTC().Format("2006-01-02 15:04:05")
	endTime := end.UTC().Format("2006-01-02 15:04:05")

	type visitorProgress struct {
		earliestTime time.Time
	}

	activeVisitors := make(map[string]visitorProgress)

	step0 := steps[0]
	step0Name := step0.Path
	if step0Name == "" {
		step0Name = step0.Name
	}

	var step0Query string
	var step0Args []interface{}

	if strings.ToLower(step0.Type) == "pageview" {
		step0Query = `
			SELECT visitor_id, MIN(created_at) as first_time
			FROM events
			WHERE site_id = ? AND event_name = '$pageview' AND path = ? AND created_at >= ? AND created_at <= ?
			GROUP BY visitor_id
		`
		step0Args = []interface{}{siteID, step0.Path, startTime, endTime}
	} else {
		step0Query = `
			SELECT visitor_id, MIN(created_at) as first_time
			FROM events
			WHERE site_id = ? AND event_name = ? AND created_at >= ? AND created_at <= ?
			GROUP BY visitor_id
		`
		step0Args = []interface{}{siteID, step0.Name, startTime, endTime}
	}

	rows, err := f.db.Query(step0Query, step0Args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query step 0: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var vID string
		var rawTime interface{}
		if err := rows.Scan(&vID, &rawTime); err != nil {
			return nil, err
		}
		activeVisitors[vID] = visitorProgress{earliestTime: parseAnyTime(rawTime)}
	}

	results := make([]*FunnelStepResult, len(steps))
	initialCount := int64(len(activeVisitors))

	results[0] = &FunnelStepResult{
		StepIndex:      0,
		Name:           step0Name,
		Type:           step0.Type,
		Visitors:       initialCount,
		ConversionRate: 100.0,
		DropOffRate:    0.0,
	}

	prevCount := initialCount

	// For subsequent steps, check if the active visitors reached the next step after their previous timestamp
	for i := 1; i < len(steps); i++ {
		step := steps[i]
		stepName := step.Path
		if stepName == "" {
			stepName = step.Name
		}

		if len(activeVisitors) == 0 {
			results[i] = &FunnelStepResult{
				StepIndex:      i,
				Name:           stepName,
				Type:           step.Type,
				Visitors:       0,
				ConversionRate: 0.0,
				DropOffRate:    100.0,
			}
			continue
		}

		nextVisitors := make(map[string]visitorProgress)

		var query string
		var args []interface{}

		if strings.ToLower(step.Type) == "pageview" {
			query = `
				SELECT visitor_id, created_at
				FROM events
				WHERE site_id = ? AND event_name = '$pageview' AND path = ? AND created_at >= ? AND created_at <= ?
				ORDER BY created_at ASC
			`
			args = []interface{}{siteID, step.Path, startTime, endTime}
		} else {
			query = `
				SELECT visitor_id, created_at
				FROM events
				WHERE site_id = ? AND event_name = ? AND created_at >= ? AND created_at <= ?
				ORDER BY created_at ASC
			`
			args = []interface{}{siteID, step.Name, startTime, endTime}
		}

		stepRows, err := f.db.Query(query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to query step %d: %w", i, err)
		}

		for stepRows.Next() {
			var vID string
			var rawTime interface{}
			if err := stepRows.Scan(&vID, &rawTime); err != nil {
				stepRows.Close()
				return nil, err
			}
			eventTime := parseAnyTime(rawTime)

			if prev, ok := activeVisitors[vID]; ok {
				if eventTime.After(prev.earliestTime) || eventTime.Equal(prev.earliestTime) {
					if existing, has := nextVisitors[vID]; !has || eventTime.Before(existing.earliestTime) {
						nextVisitors[vID] = visitorProgress{earliestTime: eventTime}
					}
				}
			}
		}
		stepRows.Close()

		currentCount := int64(len(nextVisitors))
		convRate := 0.0
		dropRate := 100.0
		if prevCount > 0 {
			convRate = (float64(currentCount) / float64(prevCount)) * 100.0
			dropRate = 100.0 - convRate
		}

		results[i] = &FunnelStepResult{
			StepIndex:      i,
			Name:           stepName,
			Type:           step.Type,
			Visitors:       currentCount,
			ConversionRate: convRate,
			DropOffRate:    dropRate,
		}

		activeVisitors = nextVisitors
		prevCount = currentCount
	}

	overallConv := 0.0
	var finalCount int64
	if len(results) > 0 {
		finalCount = results[len(results)-1].Visitors
		if initialCount > 0 {
			overallConv = (float64(finalCount) / float64(initialCount)) * 100.0
		}
	}

	return &FunnelResult{
		Steps:                 results,
		OverallConversionRate: overallConv,
		TotalStarted:          initialCount,
		TotalCompleted:        finalCount,
	}, nil
}
