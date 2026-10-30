import React, { useState, useEffect } from 'react';
import { WebVitalsReport } from '../types/api';
import { Gauge, CheckCircle2, AlertTriangle, XCircle, Sparkles, RefreshCw } from 'lucide-react';

interface WebVitalsTabProps {
  onFetchVitals: (range: string) => Promise<WebVitalsReport>;
  timeRange: string;
}

export const WebVitalsTab: React.FC<WebVitalsTabProps> = ({ onFetchVitals, timeRange }) => {
  const [report, setReport] = useState<WebVitalsReport | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await onFetchVitals(timeRange);
      setReport(res);
    } catch (err: any) {
      setError(err.message || 'Failed to load web vitals');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [timeRange]);

  const getRatingBadge = (rating: string) => {
    switch (rating) {
      case 'good':
        return (
          <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded-full border border-emerald-500/20">
            <CheckCircle2 className="w-3 h-3" /> Good
          </span>
        );
      case 'needs_improvement':
        return (
          <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-amber-400 bg-amber-500/10 px-2 py-0.5 rounded-full border border-amber-500/20">
            <AlertTriangle className="w-3 h-3" /> Needs Improvement
          </span>
        );
      case 'poor':
        return (
          <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-rose-400 bg-rose-500/10 px-2 py-0.5 rounded-full border border-rose-500/20">
            <XCircle className="w-3 h-3" /> Poor
          </span>
        );
      default:
        return null;
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <div className="flex items-center space-x-2">
              <Gauge className="w-5 h-5 text-blue-400" />
              <h2 className="text-base font-bold text-white">Core Web Vitals (p75 Benchmark)</h2>
            </div>
            <p className="text-xs text-slate-400 mt-1">
              Field performance metrics observed directly from real user browsers using the 75th percentile.
            </p>
          </div>

          <div className="flex items-center gap-3">
            <div className="flex items-center space-x-1.5 text-xs text-slate-400 bg-slate-800 px-3 py-1.5 rounded-lg border border-slate-700">
              <Sparkles className="w-3.5 h-3.5 text-blue-400" />
              <span>Timeframe: {timeRange}</span>
            </div>
            <button
              onClick={loadData}
              className="p-1.5 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors"
              title="Refresh Vitals"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
          </div>
        </div>
      </div>

      {error && (
        <div className="bg-red-500/10 border border-red-500/20 text-red-400 p-4 rounded-xl text-xs">
          {error}
        </div>
      )}

      {/* 3 Metric Cards: LCP, INP, CLS */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* LCP Card */}
        <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-5 shadow-sm space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">
              Largest Contentful Paint (LCP)
            </span>
            {report && getRatingBadge(report.lcp.rating)}
          </div>
          <div className="text-3xl font-extrabold text-white tracking-tight">
            {report ? `${(report.lcp.p75 / 1000).toFixed(2)}s` : '0.00s'}
          </div>
          <p className="text-[11px] text-slate-400">
            Measures loading performance. Google benchmark: <span className="text-emerald-400 font-medium">≤ 2.5s</span>.
          </p>
          <div className="text-[10px] text-slate-500 font-mono">
            Median: {report ? `${(report.lcp.median / 1000).toFixed(2)}s` : '0s'} &bull; {report?.lcp.count || 0} samples
          </div>
        </div>

        {/* INP Card */}
        <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-5 shadow-sm space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">
              Interaction to Next Paint (INP)
            </span>
            {report && getRatingBadge(report.inp.rating)}
          </div>
          <div className="text-3xl font-extrabold text-white tracking-tight">
            {report ? `${report.inp.p75.toFixed(0)}ms` : '0ms'}
          </div>
          <p className="text-[11px] text-slate-400">
            Measures responsiveness to user clicks and taps. Google benchmark: <span className="text-emerald-400 font-medium">≤ 200ms</span>.
          </p>
          <div className="text-[10px] text-slate-500 font-mono">
            Median: {report ? `${report.inp.median.toFixed(0)}ms` : '0ms'} &bull; {report?.inp.count || 0} samples
          </div>
        </div>

        {/* CLS Card */}
        <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-5 shadow-sm space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">
              Cumulative Layout Shift (CLS)
            </span>
            {report && getRatingBadge(report.cls.rating)}
          </div>
          <div className="text-3xl font-extrabold text-white tracking-tight">
            {report ? report.cls.p75.toFixed(3) : '0.000'}
          </div>
          <p className="text-[11px] text-slate-400">
            Measures visual layout stability. Google benchmark: <span className="text-emerald-400 font-medium">≤ 0.1</span>.
          </p>
          <div className="text-[10px] text-slate-500 font-mono">
            Median: {report ? report.cls.median.toFixed(3) : '0'} &bull; {report?.cls.count || 0} samples
          </div>
        </div>
      </div>

      {/* Pages Breakdown Table */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-xl overflow-hidden shadow-sm">
        <div className="p-5 border-b border-slate-800">
          <h3 className="text-sm font-bold text-white">Performance by Route & URL</h3>
          <p className="text-xs text-slate-400 mt-0.5">
            Average Web Vitals metrics across visited page paths.
          </p>
        </div>

        {(!report?.pages || report.pages.length === 0) ? (
          <div className="py-12 text-center text-xs text-slate-500">
            No Core Web Vitals samples recorded yet for this timeframe.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-slate-950/60 text-slate-400 border-b border-slate-800 uppercase tracking-wider text-[10px]">
                <tr>
                  <th className="py-3 px-4">Route Path</th>
                  <th className="py-3 px-4">Avg LCP</th>
                  <th className="py-3 px-4">Avg INP</th>
                  <th className="py-3 px-4">Avg CLS</th>
                  <th className="py-3 px-4 text-right">Samples</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60 text-slate-300">
                {report.pages.map((p, idx) => (
                  <tr key={idx} className="hover:bg-slate-800/30 transition-colors">
                    <td className="py-3 px-4 font-mono font-medium text-slate-200">
                      {p.path}
                    </td>
                    <td className="py-3 px-4 font-mono">
                      <span className={p.lcp <= 2500 ? 'text-emerald-400' : p.lcp <= 4000 ? 'text-amber-400' : 'text-rose-400'}>
                        {(p.lcp / 1000).toFixed(2)}s
                      </span>
                    </td>
                    <td className="py-3 px-4 font-mono">
                      <span className={p.inp <= 200 ? 'text-emerald-400' : p.inp <= 500 ? 'text-amber-400' : 'text-rose-400'}>
                        {p.inp.toFixed(0)}ms
                      </span>
                    </td>
                    <td className="py-3 px-4 font-mono">
                      <span className={p.cls <= 0.1 ? 'text-emerald-400' : p.cls <= 0.25 ? 'text-amber-400' : 'text-rose-400'}>
                        {p.cls.toFixed(3)}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-right font-mono text-slate-400">
                      {p.count}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};
