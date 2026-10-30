import React, { useState } from 'react';
import { SessionItem } from '../types/api';
import { Video, Clock, Layers, Globe, Monitor, Smartphone, AlertOctagon, Filter } from 'lucide-react';

interface SessionTableProps {
  sessions: SessionItem[];
  total: number;
  onSelectReplay: (sessionId: string) => void;
  onFilterRage?: (rageOnly: boolean) => void;
}

export const SessionTable: React.FC<SessionTableProps> = ({
  sessions,
  total,
  onSelectReplay,
  onFilterRage,
}) => {
  const [rageFilter, setRageFilter] = useState<boolean>(false);

  const toggleRage = () => {
    const next = !rageFilter;
    setRageFilter(next);
    onFilterRage?.(next);
  };

  const formatDuration = (sec: number) => {
    if (sec < 60) return `${sec}s`;
    const m = Math.floor(sec / 60);
    const s = sec % 60;
    return `${m}m ${s}s`;
  };

  return (
    <div className="bg-slate-900/80 border border-slate-800 rounded-xl overflow-hidden shadow-sm">
      <div className="p-5 border-b border-slate-800 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-base font-bold text-white flex items-center gap-2">
            <Video className="w-4 h-4 text-blue-400" />
            User Sessions & Replays
          </h2>
          <p className="text-xs text-slate-400 mt-0.5">
            Individual user journeys and DOM mutation recordings with automatic PII masking.
          </p>
        </div>

        <div className="flex items-center space-x-3">
          {/* Rage Click Filter Toggle */}
          <button
            onClick={toggleRage}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors border ${
              rageFilter
                ? 'bg-rose-500/20 text-rose-300 border-rose-500/40 shadow-sm'
                : 'bg-slate-800/80 hover:bg-slate-800 text-slate-400 border-slate-700'
            }`}
          >
            <AlertOctagon className="w-3.5 h-3.5 text-rose-400" />
            <span>Rage Clicks Only</span>
          </button>

          <span className="text-xs text-slate-400 bg-slate-800 px-2.5 py-1.5 rounded-lg font-mono border border-slate-700">
            {total} sessions
          </span>
        </div>
      </div>

      {sessions.length === 0 ? (
        <div className="py-12 text-center text-xs text-slate-500">
          No sessions recorded yet for this view.
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-slate-950/60 text-slate-400 border-b border-slate-800 font-semibold uppercase tracking-wider text-[10px]">
              <tr>
                <th className="py-3 px-4">Visitor / Session</th>
                <th className="py-3 px-4">Signals</th>
                <th className="py-3 px-4">Duration</th>
                <th className="py-3 px-4">Events</th>
                <th className="py-3 px-4">Journey Path</th>
                <th className="py-3 px-4">Environment</th>
                <th className="py-3 px-4">Location</th>
                <th className="py-3 px-4 text-right">Replay</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60 text-slate-300">
              {sessions.map((s) => (
                <tr key={s.id} className="hover:bg-slate-800/30 transition-colors">
                  <td className="py-3 px-4">
                    <div className="font-mono text-slate-200 text-xs font-medium">
                      {s.id.length > 18 ? s.id.slice(0, 18) + '...' : s.id}
                    </div>
                    <div className="text-[10px] text-slate-500 font-mono mt-0.5">
                      {new Date(s.started_at).toLocaleString()}
                    </div>
                  </td>
                  <td className="py-3 px-4">
                    {s.has_rage_click ? (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold bg-rose-500/15 text-rose-400 border border-rose-500/30">
                        <AlertOctagon className="w-3 h-3 text-rose-400" />
                        Rage Click
                      </span>
                    ) : (
                      <span className="text-[10px] text-slate-600 font-mono">-</span>
                    )}
                  </td>
                  <td className="py-3 px-4">
                    <span className="inline-flex items-center gap-1 font-mono text-slate-300">
                      <Clock className="w-3 h-3 text-slate-500" />
                      {formatDuration(s.duration_seconds)}
                    </span>
                  </td>
                  <td className="py-3 px-4">
                    <span className="inline-flex items-center gap-1 font-mono text-slate-300">
                      <Layers className="w-3 h-3 text-slate-500" />
                      {s.event_count}
                    </span>
                  </td>
                  <td className="py-3 px-4 max-w-xs truncate">
                    <div className="flex items-center space-x-1 font-mono text-xs">
                      <span className="text-blue-400">{s.entry_path || '/'}</span>
                      {s.exit_path && s.exit_path !== s.entry_path && (
                        <>
                          <span className="text-slate-600">→</span>
                          <span className="text-slate-400">{s.exit_path}</span>
                        </>
                      )}
                    </div>
                  </td>
                  <td className="py-3 px-4">
                    <div className="flex items-center space-x-1.5 text-slate-400">
                      {s.device === 'mobile' ? (
                        <Smartphone className="w-3.5 h-3.5 text-amber-400" />
                      ) : (
                        <Monitor className="w-3.5 h-3.5 text-blue-400" />
                      )}
                      <span>
                        {s.browser} on {s.os}
                      </span>
                    </div>
                  </td>
                  <td className="py-3 px-4">
                    <span className="inline-flex items-center gap-1 text-slate-400 font-mono">
                      <Globe className="w-3 h-3 text-slate-500" />
                      {s.country || 'Unknown'}
                    </span>
                  </td>
                  <td className="py-3 px-4 text-right">
                    {s.has_replay ? (
                      <button
                        onClick={() => onSelectReplay(s.id)}
                        className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 hover:text-blue-300 border border-blue-500/30 font-medium transition-colors"
                      >
                        <Video className="w-3.5 h-3.5" />
                        Watch
                      </button>
                    ) : (
                      <span className="text-[11px] text-slate-600 italic">No recording</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};
