import React from 'react';
import { RealtimeStats } from '../types/api';
import { Radio, Eye, Globe2, Activity } from 'lucide-react';

interface RealtimeCardProps {
  stats: RealtimeStats | null;
  liveEvents: any[];
}

export const RealtimeCard: React.FC<RealtimeCardProps> = ({ stats, liveEvents }) => {
  const activeCount = stats?.active_visitors || 0;
  const activePages = stats?.active_pages || [];
  const countries = stats?.countries || [];

  return (
    <div className="space-y-6">
      {/* Top Banner */}
      <div className="bg-gradient-to-br from-slate-900 via-slate-900 to-blue-950/40 border border-slate-800 rounded-xl p-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="flex items-center space-x-4">
            <div className="relative flex items-center justify-center w-12 h-12 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
              <Radio className="w-6 h-6 animate-pulse" />
              <span className="absolute -top-1 -right-1 w-3 h-3 bg-emerald-500 rounded-full animate-ping opacity-75"></span>
            </div>
            <div>
              <div className="text-sm font-medium text-slate-400">Active Visitors Right Now</div>
              <div className="text-3xl font-extrabold text-white tracking-tight flex items-baseline gap-2">
                {activeCount}
                <span className="text-xs font-normal text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded-full border border-emerald-500/20">
                  Live (last 5 min)
                </span>
              </div>
            </div>
          </div>
          <div className="text-xs text-slate-400 bg-slate-800/80 px-3 py-2 rounded-lg border border-slate-700/60 max-w-xs">
            Cookieless daily salt hashing active. Zero IP storage. Streaming via Server-Sent Events.
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Active Pages */}
        <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-5">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center space-x-2">
              <Eye className="w-4 h-4 text-blue-400" />
              <h3 className="text-sm font-semibold text-white">Active Pages</h3>
            </div>
            <span className="text-xs text-slate-500">{activePages.length} active routes</span>
          </div>

          {activePages.length === 0 ? (
            <div className="py-8 text-center text-xs text-slate-500">
              No concurrent visitors on tracked paths right now
            </div>
          ) : (
            <div className="space-y-3">
              {activePages.map((page, idx) => (
                <div key={idx} className="flex items-center justify-between text-xs">
                  <span className="font-mono text-slate-300 truncate max-w-[280px] bg-slate-800/60 px-2 py-1 rounded">
                    {page.path}
                  </span>
                  <span className="font-semibold text-blue-400">{page.views} visitors</span>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Active Countries */}
        <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-5">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center space-x-2">
              <Globe2 className="w-4 h-4 text-cyan-400" />
              <h3 className="text-sm font-semibold text-white">Active Geographies</h3>
            </div>
            <span className="text-xs text-slate-500">{countries.length} locations</span>
          </div>

          {countries.length === 0 ? (
            <div className="py-8 text-center text-xs text-slate-500">
              Waiting for incoming visitor traffic...
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-2">
              {countries.map((c, idx) => (
                <div
                  key={idx}
                  className="flex items-center justify-between text-xs bg-slate-800/40 border border-slate-800 px-3 py-2 rounded-lg"
                >
                  <span className="font-medium text-slate-300 flex items-center gap-1.5">
                    <span className="text-sm">🌐</span> {c.name}
                  </span>
                  <span className="text-cyan-400 font-semibold">{c.count}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Live Event Stream Ticker */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-5">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-2">
            <Activity className="w-4 h-4 text-amber-400" />
            <h3 className="text-sm font-semibold text-white">Live Event Stream (SSE)</h3>
          </div>
          <span className="text-xs text-slate-400 font-mono">
            {liveEvents.length > 0 ? `${liveEvents.length} events caught` : 'Listening on /api/realtime/stream'}
          </span>
        </div>

        {liveEvents.length === 0 && (!stats?.recent_events || stats.recent_events.length === 0) ? (
          <div className="py-8 text-center text-xs text-slate-500">
            Realtime feed ready. Events will appear here instantly as users click or navigate.
          </div>
        ) : (
          <div className="space-y-2 max-h-72 overflow-y-auto pr-1">
            {(liveEvents.length > 0 ? liveEvents : stats?.recent_events || []).map((ev: any, idx: number) => (
              <div
                key={idx}
                className="flex items-center justify-between p-2.5 rounded-lg bg-slate-950/60 border border-slate-800/80 text-xs hover:border-slate-700 transition-colors"
              >
                <div className="flex items-center space-x-3">
                  <span
                    className={`px-2 py-0.5 rounded text-[11px] font-mono font-medium ${
                      ev.event_name === '$pageview' || ev.event === '$pageview'
                        ? 'bg-blue-500/15 text-blue-400 border border-blue-500/30'
                        : ev.event_name?.startsWith('click') || ev.event?.startsWith('click')
                        ? 'bg-amber-500/15 text-amber-400 border border-amber-500/30'
                        : 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30'
                    }`}
                  >
                    {ev.event_name || ev.event || '$event'}
                  </span>
                  <span className="font-mono text-slate-300 truncate max-w-[240px]">
                    {ev.path || '/'}
                  </span>
                </div>
                <div className="flex items-center space-x-3 text-slate-500">
                  {ev.country && <span className="font-mono text-slate-400">[{ev.country}]</span>}
                  {ev.browser && <span>{ev.browser}</span>}
                  <span className="font-mono text-[10px]">
                    {ev.created_at ? new Date(ev.created_at).toLocaleTimeString() : 'just now'}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
