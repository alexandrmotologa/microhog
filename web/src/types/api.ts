export interface SummaryMetrics {
  pageviews: number;
  unique_visitors: number;
  sessions: number;
  bounce_rate: number;
  avg_duration_seconds: number;
}

export interface TimeseriesPoint {
  timestamp: string;
  views: number;
  visitors: number;
}

export interface PageStat {
  path: string;
  views: number;
  visitors?: number;
  percentage?: number;
}

export interface ReferrerStat {
  referrer: string;
  count: number;
  percentage?: number;
}

export interface BreakdownStat {
  name: string;
  count: number;
  percentage?: number;
}

export interface MetricsPayload {
  summary: SummaryMetrics;
  timeseries: TimeseriesPoint[];
  top_pages: PageStat[];
  top_referrers: ReferrerStat[];
  browsers: BreakdownStat[];
  os: BreakdownStat[];
  devices: BreakdownStat[];
  countries: BreakdownStat[];
  range: string;
  start: string;
  end: string;
}

export interface EventItem {
  id: number;
  site_id: string;
  session_id: string;
  visitor_id: string;
  event_name: string;
  url: string;
  path: string;
  referrer: string;
  referrer_domain: string;
  browser: string;
  os: string;
  device: string;
  country: string;
  properties_json: string;
  created_at: string;
}

export interface RealtimeStats {
  active_visitors: number;
  active_pages: PageStat[];
  countries: BreakdownStat[];
  recent_events: EventItem[];
}

export interface FunnelStepDef {
  type: 'pageview' | 'event';
  path?: string;
  name?: string;
}

export interface FunnelStepResult {
  step_index: number;
  name: string;
  type: string;
  visitors: number;
  conversion_rate: number;
  drop_off_rate: number;
}

export interface FunnelResult {
  steps: FunnelStepResult[];
  overall_conversion_rate: number;
  total_started: number;
  total_completed: number;
}

export interface SavedFunnel {
  id: string;
  site_id: string;
  name: string;
  steps_json: string;
  created_at: string;
}

export interface SessionItem {
  id: string;
  site_id: string;
  visitor_id: string;
  started_at: string;
  last_seen_at: string;
  duration_seconds: number;
  event_count: number;
  entry_path: string;
  exit_path: string;
  browser: string;
  os: string;
  device: string;
  country: string;
  has_replay: boolean;
  has_rage_click: boolean;
}

export interface VitalMetric {
  name: string;
  p75: number;
  median: number;
  rating: 'good' | 'needs_improvement' | 'poor';
  count: number;
}

export interface PageVitals {
  path: string;
  lcp: number;
  inp: number;
  cls: number;
  count: number;
}

export interface WebVitalsReport {
  lcp: VitalMetric;
  inp: VitalMetric;
  cls: VitalMetric;
  pages: PageVitals[];
}

export interface SiteItem {
  id: string;
  name: string;
  domain: string;
  created_at: string;
}

export interface WebhookItem {
  id: string;
  site_id: string;
  name: string;
  url: string;
  event_filter: string;
  created_at: string;
}
