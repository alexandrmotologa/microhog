import { useState, useEffect, useCallback, useRef } from 'react';
import { MetricsPayload, RealtimeStats, SessionItem, FunnelResult, FunnelStepDef } from '../types/api';

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
  const fetchSessions = useCallback(async (limit = 50, offset = 0) => {
    try {
      const res = await fetch(`/api/sessions?site_id=${encodeURIComponent(siteId)}&limit=${limit}&offset=${offset}`);
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
        // Incrementally refresh realtime stats
        fetchRealtime();
      } catch {}
    });

    eventSource.onerror = () => {
      // EventSource auto reconnects
    };

    return () => {
      eventSource.close();
    };
  }, [siteId, fetchRealtime]);

  // 5. Calculate Funnel
  const calculateFunnel = async (steps: FunnelStepDef[], range: string): Promise<FunnelResult> => {
    const res = await fetch('/api/funnels', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        site_id: siteId,
        steps: steps,
        range: range,
      }),
    });
    if (!res.ok) {
      const text = await res.text();
      throw new Error(text || 'Failed to calculate funnel');
    }
    return res.json();
  };

  // 6. Trigger Seed
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
    triggerSeed,
  };
}
