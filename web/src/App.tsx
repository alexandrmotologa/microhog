import React, { useState } from 'react';
import { useAnalytics } from './hooks/useAnalytics';
import { MetricCard } from './components/MetricCard';
import { TrendsChart } from './components/TrendsChart';
import { RealtimeCard } from './components/RealtimeCard';
import { FunnelChart } from './components/FunnelChart';
import { SessionTable } from './components/SessionTable';
import { ReplayModal } from './components/ReplayModal';
import { SetupGuide } from './components/SetupGuide';
import { WebVitalsTab } from './components/WebVitalsTab';
import { EventExplorer } from './components/EventExplorer';
import { SiteModal } from './components/SiteModal';
import {
  BarChart3,
  Radio,
  Filter,
  Video,
  Code2,
  RefreshCw,
  Database,
  Eye,
  Users,
  Timer,
  TrendingDown,
  Layers,
  Sparkles,
  ExternalLink,
  Gauge,
  Search,
  Download,
  Globe,
  ChevronDown,
} from 'lucide-react';

export function App() {
  const [siteId, setSiteId] = useState<string>('default');
  const [timeRange, setTimeRange] = useState<string>('7d');
  const [activeTab, setActiveTab] = useState<
    'overview' | 'realtime' | 'funnels' | 'replays' | 'vitals' | 'explorer' | 'setup'
  >('overview');
  const [selectedReplayId, setSelectedReplayId] = useState<string | null>(null);
  const [siteModalOpen, setSiteModalOpen] = useState<boolean>(false);
  const [exportOpen, setExportOpen] = useState<boolean>(false);
  const [seedNotice, setSeedNotice] = useState<string | null>(null);

  const {
    metrics,
    realtime,
    sessions,
    sessionsTotal,
    liveEvents,
    loading,
    refresh,
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
  } = useAnalytics(siteId, timeRange);

  const handleSeed = async () => {
    try {
      setSeedNotice('Seeding demo data across past 7 days...');
      await triggerSeed();
      setSeedNotice('Sample sessions, funnels, web vitals, and replays generated!');
      setTimeout(() => setSeedNotice(null), 4000);
    } catch {
      setSeedNotice('Failed to generate sample data');
      setTimeout(() => setSeedNotice(null), 3000);
    }
  };

  const formatNumber = (num: number = 0) => {
    if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
    if (num >= 1000) return (num / 1000).toFixed(1) + 'k';
    return num.toString();
  };

  const formatDuration = (sec: number = 0) => {
    if (sec < 60) return `${sec}s`;
    const m = Math.floor(sec / 60);
    const s = sec % 60;
    return `${m}m ${s}s`;
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
      {/* Top Navbar */}
      <header className="sticky top-0 z-40 bg-slate-950/80 backdrop-blur-md border-b border-slate-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between gap-3">
          {/* Logo & Product Name */}
          <div className="flex items-center space-x-3">
            <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-500 flex items-center justify-center shadow-lg shadow-blue-500/20 text-white font-black text-lg">
              µ
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-bold text-white tracking-tight text-base">MicroHog</span>
                {/* Site Switcher Trigger */}
                <button
                  onClick={() => setSiteModalOpen(true)}
                  className="flex items-center gap-1 text-[11px] font-mono px-2 py-0.5 rounded-md bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 transition-colors"
                >
                  <Globe className="w-3 h-3 text-blue-400" />
                  <span>{siteId}</span>
                  <ChevronDown className="w-3 h-3 text-slate-500" />
                </button>
              </div>
            </div>
          </div>

          {/* Center Navigation Tabs */}
          <nav className="hidden lg:flex items-center space-x-1 bg-slate-900/90 border border-slate-800 p-1 rounded-xl text-xs font-medium">
            <button
              onClick={() => setActiveTab('overview')}
              className={`flex items-center gap-1 px-2.5 py-1.5 rounded-lg transition-all ${
                activeTab === 'overview'
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'
              }`}
            >
              <BarChart3 className="w-3.5 h-3.5" />
              Overview
            </button>
            <button
              onClick={() => setActiveTab('realtime')}
              className={`flex items-center gap-1 px-2.5 py-1.5 rounded-lg transition-all ${
                activeTab === 'realtime'
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'
              }`}
            >
              <Radio className="w-3.5 h-3.5" />
              Real-Time
              {realtime && realtime.active_visitors > 0 && (
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse ml-0.5"></span>
              )}
            </button>
            <button
              onClick={() => setActiveTab('funnels')}
              className={`flex items-center gap-1 px-2.5 py-1.5 rounded-lg transition-all ${
                activeTab === 'funnels'
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'
              }`}
            >
              <Filter className="w-3.5 h-3.5" />
              Funnels
            </button>
            <button
              onClick={() => setActiveTab('replays')}
              className={`flex items-center gap-1 px-2.5 py-1.5 rounded-lg transition-all ${
                activeTab === 'replays'
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'
              }`}
            >
              <Video className="w-3.5 h-3.5" />
              Replays
            </button>
            <button
              onClick={() => setActiveTab('vitals')}
              className={`flex items-center gap-1 px-2.5 py-1.5 rounded-lg transition-all ${
                activeTab === 'vitals'
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'
              }`}
            >
              <Gauge className="w-3.5 h-3.5" />
              Web Vitals
            </button>
            <button
              onClick={() => setActiveTab('explorer')}
              className={`flex items-center gap-1 px-2.5 py-1.5 rounded-lg transition-all ${
                activeTab === 'explorer'
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'
              }`}
            >
              <Search className="w-3.5 h-3.5" />
              Events
            </button>
            <button
              onClick={() => setActiveTab('setup')}
              className={`flex items-center gap-1 px-2.5 py-1.5 rounded-lg transition-all ${
                activeTab === 'setup'
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'
              }`}
            >
              <Code2 className="w-3.5 h-3.5" />
              Settings
            </button>
          </nav>

          {/* Right Controls */}
          <div className="flex items-center space-x-2">
            {/* Live Indicator Pill */}
            <div className="flex items-center space-x-1.5 px-2.5 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-medium">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-ping"></span>
              <span>{realtime?.active_visitors || 0} online</span>
            </div>

            {/* Time range selector */}
            <select
              value={timeRange}
              onChange={(e) => setTimeRange(e.target.value)}
              className="bg-slate-900 border border-slate-700 text-slate-300 text-xs rounded-lg px-2.5 py-1.5 outline-none font-medium cursor-pointer"
            >
              <option value="24h">24h</option>
              <option value="7d">7d</option>
              <option value="30d">30d</option>
              <option value="90d">90d</option>
            </select>

            {/* Export Dropdown */}
            <div className="relative">
              <button
                onClick={() => setExportOpen(!exportOpen)}
                className="flex items-center gap-1 text-xs px-2.5 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white border border-slate-700 transition-colors"
                title="Export Data"
              >
                <Download className="w-3.5 h-3.5 text-blue-400" />
                <span className="hidden sm:inline">Export</span>
              </button>

              {exportOpen && (
                <div className="absolute right-0 mt-1 w-44 bg-slate-900 border border-slate-800 rounded-xl shadow-xl p-1 z-50 text-xs text-slate-300 space-y-0.5">
                  <a
                    href={`/api/export/events?site_id=${encodeURIComponent(siteId)}&range=${timeRange}&format=csv`}
                    download
                    onClick={() => setExportOpen(false)}
                    className="block px-3 py-1.5 rounded-lg hover:bg-slate-800 hover:text-white transition-colors"
                  >
                    Events (CSV)
                  </a>
                  <a
                    href={`/api/export/events?site_id=${encodeURIComponent(siteId)}&range=${timeRange}&format=json`}
                    download
                    onClick={() => setExportOpen(false)}
                    className="block px-3 py-1.5 rounded-lg hover:bg-slate-800 hover:text-white transition-colors"
                  >
                    Events (JSON)
                  </a>
                  <a
                    href={`/api/export/sessions?site_id=${encodeURIComponent(siteId)}&range=${timeRange}&format=csv`}
                    download
                    onClick={() => setExportOpen(false)}
                    className="block px-3 py-1.5 rounded-lg hover:bg-slate-800 hover:text-white transition-colors"
                  >
                    Sessions (CSV)
                  </a>
                  <a
                    href={`/api/export/sessions?site_id=${encodeURIComponent(siteId)}&range=${timeRange}&format=json`}
                    download
                    onClick={() => setExportOpen(false)}
                    className="block px-3 py-1.5 rounded-lg hover:bg-slate-800 hover:text-white transition-colors"
                  >
                    Sessions (JSON)
                  </a>
                </div>
              )}
            </div>

            {/* Seed Demo Button */}
            <button
              onClick={handleSeed}
              title="Generate synthetic traffic, funnels, and session replays"
              className="hidden sm:flex items-center gap-1 text-xs px-2.5 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white border border-slate-700 transition-colors"
            >
              <Database className="w-3.5 h-3.5 text-blue-400" />
              <span>Seed</span>
            </button>

            {/* Refresh */}
            <button
              onClick={refresh}
              className="p-1.5 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors"
              title="Refresh statistics"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
          </div>
        </div>

        {/* Mobile Sub-Navigation */}
        <div className="flex lg:hidden items-center space-x-1.5 border-t border-slate-800/80 px-3 py-2 bg-slate-950/95 overflow-x-auto text-xs scrollbar-none">
          <button
            onClick={() => setActiveTab('overview')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg whitespace-nowrap transition-all ${
              activeTab === 'overview'
                ? 'bg-blue-600 text-white font-semibold shadow-sm'
                : 'text-slate-400 hover:text-slate-200 bg-slate-900/60 border border-slate-800/60'
            }`}
          >
            <BarChart3 className="w-3.5 h-3.5" />
            <span>Overview</span>
          </button>
          <button
            onClick={() => setActiveTab('realtime')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg whitespace-nowrap transition-all ${
              activeTab === 'realtime'
                ? 'bg-blue-600 text-white font-semibold shadow-sm'
                : 'text-slate-400 hover:text-slate-200 bg-slate-900/60 border border-slate-800/60'
            }`}
          >
            <Radio className="w-3.5 h-3.5" />
            <span>Real-Time</span>
            {realtime && realtime.active_visitors > 0 && (
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
            )}
          </button>
          <button
            onClick={() => setActiveTab('funnels')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg whitespace-nowrap transition-all ${
              activeTab === 'funnels'
                ? 'bg-blue-600 text-white font-semibold shadow-sm'
                : 'text-slate-400 hover:text-slate-200 bg-slate-900/60 border border-slate-800/60'
            }`}
          >
            <Filter className="w-3.5 h-3.5" />
            <span>Funnels</span>
          </button>
          <button
            onClick={() => setActiveTab('replays')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg whitespace-nowrap transition-all ${
              activeTab === 'replays'
                ? 'bg-blue-600 text-white font-semibold shadow-sm'
                : 'text-slate-400 hover:text-slate-200 bg-slate-900/60 border border-slate-800/60'
            }`}
          >
            <Video className="w-3.5 h-3.5" />
            <span>Replays</span>
          </button>
          <button
            onClick={() => setActiveTab('vitals')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg whitespace-nowrap transition-all ${
              activeTab === 'vitals'
                ? 'bg-blue-600 text-white font-semibold shadow-sm'
                : 'text-slate-400 hover:text-slate-200 bg-slate-900/60 border border-slate-800/60'
            }`}
          >
            <Gauge className="w-3.5 h-3.5" />
            <span>Web Vitals</span>
          </button>
          <button
            onClick={() => setActiveTab('explorer')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg whitespace-nowrap transition-all ${
              activeTab === 'explorer'
                ? 'bg-blue-600 text-white font-semibold shadow-sm'
                : 'text-slate-400 hover:text-slate-200 bg-slate-900/60 border border-slate-800/60'
            }`}
          >
            <Search className="w-3.5 h-3.5" />
            <span>Events</span>
          </button>
          <button
            onClick={() => setActiveTab('setup')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg whitespace-nowrap transition-all ${
              activeTab === 'setup'
                ? 'bg-blue-600 text-white font-semibold shadow-sm'
                : 'text-slate-400 hover:text-slate-200 bg-slate-900/60 border border-slate-800/60'
            }`}
          >
            <Code2 className="w-3.5 h-3.5" />
            <span>Settings</span>
          </button>
        </div>
      </header>

      {/* Seed Notice Toast */}
      {seedNotice && (
        <div className="bg-blue-600 text-white text-xs py-2 px-4 text-center font-medium shadow-md transition-all">
          {seedNotice}
        </div>
      )}

      {/* Main Content Area */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* TAB 1: OVERVIEW */}
        {activeTab === 'overview' && (
          <div className="space-y-6">
            {/* Top Metric Cards Grid */}
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3 sm:gap-4">
              <MetricCard
                title="Pageviews"
                value={formatNumber(metrics?.summary.pageviews)}
                subtext={`Total visits in ${timeRange}`}
                icon={Eye}
                color="blue"
              />
              <MetricCard
                title="Unique Visitors"
                value={formatNumber(metrics?.summary.unique_visitors)}
                subtext="Daily salted IDs"
                icon={Users}
                color="emerald"
              />
              <MetricCard
                title="Total Sessions"
                value={formatNumber(metrics?.summary.sessions)}
                subtext="User journeys"
                icon={Layers}
                color="purple"
              />
              <MetricCard
                title="Bounce Rate"
                value={`${(metrics?.summary.bounce_rate || 0).toFixed(1)}%`}
                subtext="Single page exits"
                icon={TrendingDown}
                color="amber"
              />
              <MetricCard
                title="Avg Duration"
                value={formatDuration(metrics?.summary.avg_duration_seconds)}
                subtext="Time engaged"
                icon={Timer}
                color="blue"
              />
            </div>

            {/* Time-Series Chart */}
            <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-6 shadow-sm">
              <div className="flex items-center justify-between mb-2">
                <h2 className="text-base font-bold text-white">Traffic Trends</h2>
                <span className="text-xs text-slate-500 font-mono">
                  {timeRange === '24h' ? 'Hourly aggregation' : 'Daily aggregation'}
                </span>
              </div>
              <TrendsChart data={metrics?.timeseries || []} />
            </div>

            {/* Breakdown Tables (2 columns) */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Top Pages */}
              <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-5 shadow-sm">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-sm font-semibold text-white">Top Pages & Routes</h3>
                  <span className="text-xs text-slate-500">Views</span>
                </div>
                {!metrics?.top_pages || metrics.top_pages.length === 0 ? (
                  <div className="py-8 text-center text-xs text-slate-500">No pages recorded yet</div>
                ) : (
                  <div className="space-y-3">
                    {metrics.top_pages.map((p, idx) => (
                      <div key={idx} className="space-y-1">
                        <div className="flex items-center justify-between text-xs">
                          <span className="font-mono text-slate-200 truncate max-w-[280px]">{p.path}</span>
                          <span className="font-semibold text-slate-300">{p.views}</span>
                        </div>
                        <div className="w-full bg-slate-950 rounded-full h-1.5 overflow-hidden">
                          <div
                            className="bg-blue-500 h-full rounded-full"
                            style={{ width: `${Math.min(100, p.percentage || 10)}%` }}
                          />
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Top Referrers */}
              <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-5 shadow-sm">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-sm font-semibold text-white">Top Referrers</h3>
                  <span className="text-xs text-slate-500">Referrals</span>
                </div>
                {!metrics?.top_referrers || metrics.top_referrers.length === 0 ? (
                  <div className="py-8 text-center text-xs text-slate-500">No referrer data yet</div>
                ) : (
                  <div className="space-y-3">
                    {metrics.top_referrers.map((r, idx) => (
                      <div key={idx} className="space-y-1">
                        <div className="flex items-center justify-between text-xs">
                          <span className="font-medium text-slate-200">{r.referrer}</span>
                          <span className="font-semibold text-slate-300">{r.count}</span>
                        </div>
                        <div className="w-full bg-slate-950 rounded-full h-1.5 overflow-hidden">
                          <div
                            className="bg-emerald-500 h-full rounded-full"
                            style={{ width: `${Math.min(100, r.percentage || 10)}%` }}
                          />
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>

            {/* Demographics Grid (Devices, Browsers, Operating Systems, Countries) */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              {/* Devices */}
              <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-4">
                <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-3">
                  Devices
                </h4>
                <div className="space-y-2">
                  {metrics?.devices.map((d, i) => (
                    <div key={i} className="flex items-center justify-between text-xs">
                      <span className="capitalize text-slate-300">{d.name}</span>
                      <span className="font-mono text-slate-400">
                        {d.count} ({(d.percentage || 0).toFixed(0)}%)
                      </span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Browsers */}
              <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-4">
                <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-3">
                  Browsers
                </h4>
                <div className="space-y-2">
                  {metrics?.browsers.map((b, i) => (
                    <div key={i} className="flex items-center justify-between text-xs">
                      <span className="text-slate-300">{b.name}</span>
                      <span className="font-mono text-slate-400">
                        {b.count} ({(b.percentage || 0).toFixed(0)}%)
                      </span>
                    </div>
                  ))}
                </div>
              </div>

              {/* OS */}
              <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-4">
                <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-3">
                  Operating Systems
                </h4>
                <div className="space-y-2">
                  {metrics?.os.map((o, i) => (
                    <div key={i} className="flex items-center justify-between text-xs">
                      <span className="text-slate-300">{o.name}</span>
                      <span className="font-mono text-slate-400">
                        {o.count} ({(o.percentage || 0).toFixed(0)}%)
                      </span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Countries */}
              <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-4">
                <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-3">
                  Countries
                </h4>
                <div className="space-y-2">
                  {metrics?.countries.slice(0, 5).map((c, i) => (
                    <div key={i} className="flex items-center justify-between text-xs">
                      <span className="text-slate-300 font-mono">🌐 {c.name}</span>
                      <span className="font-mono text-slate-400">
                        {c.count} ({(c.percentage || 0).toFixed(0)}%)
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* TAB 2: REAL-TIME FEED */}
        {activeTab === 'realtime' && (
          <RealtimeCard stats={realtime} liveEvents={liveEvents} />
        )}

        {/* TAB 3: FUNNELS */}
        {activeTab === 'funnels' && (
          <FunnelChart
            onCalculate={calculateFunnel}
            onFetchSavedFunnels={fetchSavedFunnels}
            onSaveFunnel={saveFunnel}
            onDeleteSavedFunnel={deleteSavedFunnel}
            timeRange={timeRange}
          />
        )}

        {/* TAB 4: SESSION REPLAYS */}
        {activeTab === 'replays' && (
          <SessionTable
            sessions={sessions}
            total={sessionsTotal}
            onSelectReplay={(sid) => setSelectedReplayId(sid)}
            onFilterRage={(rageOnly) => fetchSessions(50, 0, rageOnly)}
          />
        )}

        {/* TAB 5: WEB VITALS */}
        {activeTab === 'vitals' && (
          <WebVitalsTab onFetchVitals={fetchWebVitals} timeRange={timeRange} />
        )}

        {/* TAB 6: EVENT EXPLORER */}
        {activeTab === 'explorer' && (
          <EventExplorer onSearchEvents={searchEvents} />
        )}

        {/* TAB 7: SETUP & SETTINGS */}
        {activeTab === 'setup' && (
          <SetupGuide
            currentSiteId={siteId}
            onFetchWebhooks={fetchWebhooks}
            onSaveWebhook={saveWebhook}
            onDeleteWebhook={deleteWebhook}
            onTestWebhook={testWebhook}
          />
        )}
      </main>

      {/* Session Replay Modal */}
      {selectedReplayId && (
        <ReplayModal sessionId={selectedReplayId} onClose={() => setSelectedReplayId(null)} />
      )}

      {/* Site Management Modal */}
      {siteModalOpen && (
        <SiteModal
          currentSiteId={siteId}
          onSelectSite={(id) => setSiteId(id)}
          onClose={() => setSiteModalOpen(false)}
          onFetchSites={fetchSites}
          onCreateSite={createSite}
          onDeleteSite={deleteSite}
        />
      )}

      {/* Footer */}
      <footer className="border-t border-slate-800/80 py-6 text-center text-xs text-slate-500">
        <div className="max-w-7xl mx-auto px-4 flex flex-col sm:flex-row items-center justify-between gap-3">
          <div>MicroHog &bull; Lightweight Self-Hosted Product Analytics &bull; MIT License</div>
          <div className="flex items-center space-x-4">
            <a
              href="https://github.com/alexandrmotologa/microhog"
              target="_blank"
              rel="noopener noreferrer"
              className="text-slate-400 hover:text-white flex items-center gap-1 transition-colors"
            >
              <span>GitHub</span>
              <ExternalLink className="w-3 h-3" />
            </a>
          </div>
        </div>
      </footer>
    </div>
  );
}
