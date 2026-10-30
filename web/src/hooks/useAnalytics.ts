import { useState, useEffect, useCallback } from 'react';
import {
  MetricsPayload,
  RealtimeStats,
  SessionItem,
  FunnelResult,
  FunnelStepDef,
  SavedFunnel,
  WebVitalsReport,
  SiteItem,
  WebhookItem,
  EventItem,
} from '../types/api';

export function useAnalytics(siteId: string, timeRange: string) {
  const [metrics, setMetrics] = useState<MetricsPayload | null>(null);
  const [realtime, setRealtime] = useState<RealtimeStats | null>(null);
  const [sessions, setSessions] = useState<SessionItem[]>([]);
  const [sessionsTotal, setSessionsTotal] = useState<number>(0);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [liveEvents, setLiveEvents] = useState<any[]>([]);

  // 1. Fetch Overview Metrics
  const fetchMetrics = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await fetch(`/api/metrics?site_id=${encodeURIComponent(siteId)}&range=${timeRange}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data: MetricsPayload = await res.json();
      setMetrics(data);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch metrics');
    } finally {
      setLoading(false);
    }
  }, [siteId, timeRange]);

  // 2. Fetch Sessions
  const fetchSessions = useCallback(async (limit = 50, offset = 0, rageOnly = false) => {
    try {
      const res = await fetch(
        `/api/sessions?site_id=${encodeURIComponent(siteId)}&limit=${limit}&offset=${offset}${rageOnly ? '&rage_only=true' : ''}`
      );
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setSessions(data.sessions || []);
      setSessionsTotal(data.total || 0);
    } catch (err) {
      console.error('Failed to load sessions', err);
    }
  }, [siteId]);

  // 3. Fetch Realtime Snapshot
  const fetchRealtime = useCallback(async () => {
    try {
      const res = await fetch(`/api/realtime?site_id=${encodeURIComponent(siteId)}`);
      if (!res.ok) return;
      const data: RealtimeStats = await res.json();
      setRealtime(data);
    } catch {}
  }, [siteId]);

  // Initial loads
  useEffect(() => {
    fetchMetrics();
    fetchSessions();
    fetchRealtime();
  }, [fetchMetrics, fetchSessions, fetchRealtime]);

  // 4. Server-Sent Events (SSE) for Real-Time Streaming
  useEffect(() => {
    const sseUrl = `/api/realtime/stream?site_id=${encodeURIComponent(siteId)}`;
    const eventSource = new EventSource(sseUrl);

    eventSource.addEventListener('snapshot', (e) => {
      try {
        const snap = JSON.parse(e.data);
        setRealtime(snap);
      } catch {}
    });

    eventSource.addEventListener('live_event', (e) => {
      try {
        const eventData = JSON.parse(e.data);
        setLiveEvents((prev) => [eventData, ...prev.slice(0, 49)]);
        fetchRealtime();
      } catch {}
    });

    eventSource.onerror = () => {};

    return () => {
      eventSource.close();
    };
  }, [siteId, fetchRealtime]);

  // 5. Funnels
  const calculateFunnel = async (steps: FunnelStepDef[], range: string): Promise<FunnelResult> => {
    const res = await fetch('/api/funnels', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ site_id: siteId, steps, range }),
    });
    if (!res.ok) {
      const text = await res.text();
      throw new Error(text || 'Failed to calculate funnel');
    }
    return res.json();
  };

  const fetchSavedFunnels = async (): Promise<SavedFunnel[]> => {
    const res = await fetch(`/api/funnels/saved?site_id=${encodeURIComponent(siteId)}`);
    if (!res.ok) return [];
    return res.json();
  };

  const saveFunnel = async (name: string, steps: FunnelStepDef[]): Promise<SavedFunnel> => {
    const res = await fetch('/api/funnels/saved', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ site_id: siteId, name, steps }),
    });
    if (!res.ok) throw new Error('Failed to save funnel');
    return res.json();
  };

  const deleteSavedFunnel = async (id: string): Promise<void> => {
    await fetch(`/api/funnels/saved/${encodeURIComponent(id)}`, { method: 'DELETE' });
  };

  // 6. Core Web Vitals
  const fetchWebVitals = async (range = '7d'): Promise<WebVitalsReport> => {
    const res = await fetch(`/api/vitals?site_id=${encodeURIComponent(siteId)}&range=${range}`);
    if (!res.ok) throw new Error('Failed to fetch web vitals');
    return res.json();
  };

  // 7. Event Explorer / Search
  const searchEvents = async (query = '', eventName = 'all', limit = 50, offset = 0): Promise<{ events: EventItem[]; total: number }> => {
    const res = await fetch(
      `/api/events/search?site_id=${encodeURIComponent(siteId)}&query=${encodeURIComponent(query)}&event_name=${encodeURIComponent(eventName)}&limit=${limit}&offset=${offset}`
    );
    if (!res.ok) throw new Error('Failed to search events');
    return res.json();
  };

  // 8. Sites Management
  const fetchSites = async (): Promise<SiteItem[]> => {
    const res = await fetch('/api/sites');
    if (!res.ok) return [];
    return res.json();
  };

  const createSite = async (id: string, name: string, domain: string): Promise<void> => {
    const res = await fetch('/api/sites', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id, name, domain }),
    });
    if (!res.ok) throw new Error('Failed to create site');
  };

  const deleteSite = async (id: string): Promise<void> => {
    const res = await fetch(`/api/sites/${encodeURIComponent(id)}`, { method: 'DELETE' });
    if (!res.ok) throw new Error('Failed to delete site');
  };

  // 9. Webhooks
  const fetchWebhooks = async (): Promise<WebhookItem[]> => {
    const res = await fetch(`/api/webhooks?site_id=${encodeURIComponent(siteId)}`);
    if (!res.ok) return [];
    return res.json();
  };

  const saveWebhook = async (name: string, url: string, eventFilter = 'all'): Promise<WebhookItem> => {
    const res = await fetch('/api/webhooks', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ site_id: siteId, name, url, event_filter: eventFilter }),
    });
    if (!res.ok) throw new Error('Failed to save webhook');
    return res.json();
  };

  const deleteWebhook = async (id: string): Promise<void> => {
    await fetch(`/api/webhooks/${encodeURIComponent(id)}`, { method: 'DELETE' });
  };

  const testWebhook = async (url: string): Promise<void> => {
    const res = await fetch('/api/webhooks/test', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url }),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.message || 'Webhook verification failed');
    }
  };

  // 10. Seed Trigger
  const triggerSeed = async () => {
    const res = await fetch(`/api/seed?site_id=${encodeURIComponent(siteId)}`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to seed demo data');
    await fetchMetrics();
    await fetchSessions();
    await fetchRealtime();
  };

  return {
    metrics,
    realtime,
    sessions,
    sessionsTotal,
    liveEvents,
    loading,
    error,
    refresh: fetchMetrics,
    fetchSessions,
    calculateFunnel,
    fetchSavedFunnels,
    saveFunnel,
    deleteSavedFunnel,
    fetchWebVitals,
    searchEvents,
    fetchSites,
    createSite,
    deleteSite,
    fetchWebhooks,
    saveWebhook,
    deleteWebhook,
    testWebhook,
    triggerSeed,
  };
}
