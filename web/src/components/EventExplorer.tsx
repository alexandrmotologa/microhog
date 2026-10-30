import React, { useState, useEffect } from 'react';
import { EventItem } from '../types/api';
import { Search, Filter, Code, X, Calendar, Globe, Monitor, RefreshCw } from 'lucide-react';

interface EventExplorerProps {
  onSearchEvents: (query: string, eventName: string) => Promise<{ events: EventItem[]; total: number }>;
}

export const EventExplorer: React.FC<EventExplorerProps> = ({ onSearchEvents }) => {
  const [events, setEvents] = useState<EventItem[]>([]);
  const [total, setTotal] = useState<number>(0);
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [selectedEventName, setSelectedEventName] = useState<string>('all');
  const [loading, setLoading] = useState<boolean>(false);
  const [inspectEvent, setInspectEvent] = useState<EventItem | null>(null);

  const performSearch = async () => {
    try {
      setLoading(true);
      const res = await onSearchEvents(searchQuery, selectedEventName);
      setEvents(res.events || []);
      setTotal(res.total || 0);
    } catch (err) {
      console.error('Search failed', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    performSearch();
  }, [selectedEventName]);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      performSearch();
    }
  };

  return (
    <div className="space-y-6">
      {/* Search Header */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-6">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h2 className="text-base font-bold text-white flex items-center gap-2">
              <Code className="w-5 h-5 text-blue-400" />
              Event Explorer & Property Inspector
            </h2>
            <p className="text-xs text-slate-400 mt-1">
              Search and inspect raw JSON event properties, metadata, and custom telemetry in real time.
            </p>
          </div>

          <span className="text-xs text-slate-400 bg-slate-800 px-3 py-1.5 rounded-lg border border-slate-700 font-mono">
            {total} matching events
          </span>
        </div>

        {/* Filter bar */}
        <div className="mt-5 flex flex-col sm:flex-row items-center gap-3">
          <div className="relative flex-1 w-full">
            <Search className="w-4 h-4 text-slate-500 absolute left-3 top-2.5" />
            <input
              type="text"
              placeholder="Search by path, property JSON, or visitor ID (press Enter)..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={handleKeyDown}
              className="w-full bg-slate-950 border border-slate-700 text-slate-200 rounded-lg pl-9 pr-3 py-2 text-xs font-mono outline-none focus:border-blue-500"
            />
          </div>

          <div className="flex items-center gap-2 w-full sm:w-auto">
            <select
              value={selectedEventName}
              onChange={(e) => setSelectedEventName(e.target.value)}
              className="bg-slate-950 border border-slate-700 text-slate-300 text-xs rounded-lg px-3 py-2 outline-none font-medium cursor-pointer"
            >
              <option value="all">All Events</option>
              <option value="$pageview">$pageview</option>
              <option value="$rage_click">$rage_click</option>
              <option value="$web_vitals">$web_vitals</option>
              <option value="$outbound_click">$outbound_click</option>
              <option value="click_signup">click_signup</option>
              <option value="signup_completed">signup_completed</option>
            </select>

            <button
              onClick={performSearch}
              className="flex items-center gap-1.5 px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-medium transition-colors"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
              <span>Search</span>
            </button>
          </div>
        </div>
      </div>

      {/* Events Table */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-xl overflow-hidden shadow-sm">
        {events.length === 0 ? (
          <div className="py-12 text-center text-xs text-slate-500">
            No events found matching your search query.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-slate-950/60 text-slate-400 border-b border-slate-800 uppercase tracking-wider text-[10px]">
                <tr>
                  <th className="py-3 px-4">Event</th>
                  <th className="py-3 px-4">Path</th>
                  <th className="py-3 px-4">Visitor</th>
                  <th className="py-3 px-4">Client</th>
                  <th className="py-3 px-4">Timestamp</th>
                  <th className="py-3 px-4 text-right">Details</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60 text-slate-300">
                {events.map((e) => (
                  <tr key={e.id} className="hover:bg-slate-800/30 transition-colors">
                    <td className="py-3 px-4">
                      <span
                        className={`inline-block px-2 py-0.5 rounded text-[11px] font-mono font-medium ${
                          e.event_name === '$pageview'
                            ? 'bg-blue-500/15 text-blue-400 border border-blue-500/30'
                            : e.event_name === '$rage_click'
                            ? 'bg-rose-500/15 text-rose-400 border border-rose-500/30 animate-pulse'
                            : e.event_name === '$web_vitals'
                            ? 'bg-purple-500/15 text-purple-400 border border-purple-500/30'
                            : 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30'
                        }`}
                      >
                        {e.event_name}
                      </span>
                    </td>
                    <td className="py-3 px-4 font-mono text-slate-200 truncate max-w-xs">
                      {e.path || '/'}
                    </td>
                    <td className="py-3 px-4 font-mono text-slate-400">
                      {e.visitor_id.slice(0, 10)}...
                    </td>
                    <td className="py-3 px-4 text-slate-400">
                      <span>
                        {e.browser} &bull; {e.country || 'US'}
                      </span>
                    </td>
                    <td className="py-3 px-4 font-mono text-slate-500 text-[11px]">
                      {new Date(e.created_at).toLocaleString()}
                    </td>
                    <td className="py-3 px-4 text-right">
                      <button
                        onClick={() => setInspectEvent(e)}
                        className="px-2.5 py-1 rounded-md bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white font-mono text-[11px] transition-colors"
                      >
                        Inspect JSON
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* JSON Inspector Drawer Modal */}
      {inspectEvent && (
        <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-2xl max-h-[85vh] flex flex-col shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-slate-950/60">
              <div className="flex items-center space-x-2">
                <Code className="w-4 h-4 text-blue-400" />
                <span className="text-sm font-bold text-white font-mono">
                  {inspectEvent.event_name}
                </span>
                <span className="text-xs text-slate-500 font-mono">
                  #{inspectEvent.id}
                </span>
              </div>
              <button
                onClick={() => setInspectEvent(null)}
                className="p-1.5 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6 overflow-y-auto space-y-4 text-xs font-mono">
              <div className="grid grid-cols-2 gap-3 bg-slate-950 p-4 rounded-xl border border-slate-800 text-slate-300">
                <div>
                  <span className="text-slate-500">Path:</span> {inspectEvent.path}
                </div>
                <div>
                  <span className="text-slate-500">Session ID:</span> {inspectEvent.session_id}
                </div>
                <div>
                  <span className="text-slate-500">Visitor ID:</span> {inspectEvent.visitor_id}
                </div>
                <div>
                  <span className="text-slate-500">Location:</span> {inspectEvent.country}
                </div>
                <div>
                  <span className="text-slate-500">Client:</span> {inspectEvent.browser} on {inspectEvent.os} ({inspectEvent.device})
                </div>
                <div>
                  <span className="text-slate-500">Timestamp:</span> {inspectEvent.created_at}
                </div>
              </div>

              <div>
                <div className="text-slate-400 font-semibold mb-1">
                  Custom Properties JSON:
                </div>
                <pre className="bg-slate-950 border border-slate-800 p-4 rounded-xl text-emerald-300 overflow-x-auto leading-relaxed">
                  {JSON.stringify(JSON.parse(inspectEvent.properties_json || '{}'), null, 2)}
                </pre>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
