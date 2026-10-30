package analytics

import (
	"fmt"
	"time"

	"github.com/alexandrmotologa/microhog/internal/storage"
)

// MetricsPayload is the comprehensive analytics report for the dashboard.
type MetricsPayload struct {
	Summary       *storage.SummaryMetrics    `json:"summary"`
	Timeseries    []*storage.TimeseriesPoint `json:"timeseries"`
	TopPages      []*storage.PageStat        `json:"top_pages"`
	TopReferrers  []*storage.ReferrerStat    `json:"top_referrers"`
	Browsers      []*storage.BreakdownStat   `json:"browsers"`
	OS            []*storage.BreakdownStat   `json:"os"`
	Devices       []*storage.BreakdownStat   `json:"devices"`
	Countries     []*storage.BreakdownStat   `json:"countries"`
	Range         string                     `json:"range"`
	Start         string                     `json:"start"`
	End           string                     `json:"end"`
}

// MetricsService bundles queries to assemble dashboard payload responses.
type MetricsService struct {
	eventStore *storage.EventStore
}

// NewMetricsService constructs a MetricsService.
func NewMetricsService(es *storage.EventStore) *MetricsService {
	return &MetricsService{eventStore: es}
}

// GetOverview builds the complete analytics overview for a given time range.
func (m *MetricsService) GetOverview(siteID, timeRange string) (*MetricsPayload, error) {
	now := time.Now().UTC()
	var start time.Time

	switch timeRange {
	case "24h":
		start = now.Add(-24 * time.Hour)
	case "7d":
		start = now.Add(-7 * 24 * time.Hour)
	case "30d":
		start = now.Add(-30 * 24 * time.Hour)
	case "90d":
		start = now.Add(-90 * 24 * time.Hour)
	default:
		timeRange = "7d"
		start = now.Add(-7 * 24 * time.Hour)
	}

	summary, err := m.eventStore.GetMetricsSummary(siteID, start, now)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch summary: %w", err)
	}

	timeseries, err := m.eventStore.GetTimeseries(siteID, start, now)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch timeseries: %w", err)
	}

	topPages, err := m.eventStore.GetTopPages(siteID, start, now, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top pages: %w", err)
	}

	topRefs, err := m.eventStore.GetTopReferrers(siteID, start, now, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top referrers: %w", err)
	}

	browsers, err := m.eventStore.GetBreakdown(siteID, "browser", start, now, 6)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch browsers: %w", err)
	}

	osStats, err := m.eventStore.GetBreakdown(siteID, "os", start, now, 6)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch os stats: %w", err)
	}

	devices, err := m.eventStore.GetBreakdown(siteID, "device", start, now, 4)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch devices: %w", err)
	}

	countries, err := m.eventStore.GetBreakdown(siteID, "country", start, now, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch countries: %w", err)
	}

	return &MetricsPayload{
		Summary:      summary,
		Timeseries:   timeseries,
		TopPages:     topPages,
		TopReferrers: topRefs,
		Browsers:     browsers,
		OS:           osStats,
		Devices:      devices,
		Countries:    countries,
		Range:        timeRange,
		Start:        start.Format(time.RFC3339),
		End:          now.Format(time.RFC3339),
	}, nil
}
