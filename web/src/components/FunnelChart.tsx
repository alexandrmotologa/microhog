import React, { useState, useEffect, useCallback } from 'react';
import { FunnelResult, FunnelStepDef } from '../types/api';
import { Filter, ArrowDown, Plus, Trash2, Play, Sparkles } from 'lucide-react';

interface FunnelChartProps {
  onCalculate: (steps: FunnelStepDef[], range: string) => Promise<FunnelResult>;
  timeRange: string;
}

const PRESET_SAAS: FunnelStepDef[] = [
  { type: 'pageview', path: '/' },
  { type: 'pageview', path: '/pricing' },
  { type: 'event', name: 'click_signup' },
  { type: 'pageview', path: '/signup' },
  { type: 'pageview', path: '/app/dashboard' },
];

const PRESET_DOCS: FunnelStepDef[] = [
  { type: 'pageview', path: '/' },
  { type: 'pageview', path: '/docs' },
  { type: 'pageview', path: '/pricing' },
];

export const FunnelChart: React.FC<FunnelChartProps> = ({ onCalculate, timeRange }) => {
  const [steps, setSteps] = useState<FunnelStepDef[]>(PRESET_SAAS);
  const [result, setResult] = useState<FunnelResult | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const runCalculation = useCallback(async (currentSteps: FunnelStepDef[]) => {
    try {
      setLoading(true);
      setError(null);
      const res = await onCalculate(currentSteps, timeRange);
      setResult(res);
    } catch (err: any) {
      setError(err.message || 'Funnel calculation failed');
    } finally {
      setLoading(false);
    }
  }, [onCalculate, timeRange]);

  useEffect(() => {
    runCalculation(steps);
  }, [runCalculation, timeRange]);

  const addStep = () => {
    setSteps([...steps, { type: 'pageview', path: '/new-step' }]);
  };

  const removeStep = (idx: number) => {
    if (steps.length <= 2) return;
    const next = steps.filter((_, i) => i !== idx);
    setSteps(next);
  };

  const updateStep = (idx: number, updated: Partial<FunnelStepDef>) => {
    const next = [...steps];
    next[idx] = { ...next[idx], ...updated };
    setSteps(next);
  };

  return (
    <div className="space-y-6">
      {/* Header & Controls */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-6">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div className="flex items-center space-x-2">
              <Filter className="w-5 h-5 text-blue-400" />
              <h2 className="text-lg font-bold text-white">Conversion Funnel Studio</h2>
            </div>
            <p className="text-xs text-slate-400 mt-1">
              Analyze multi-step user drop-offs sequentially across pages and custom events.
            </p>
          </div>

          {/* Presets */}
          <div className="flex items-center gap-2">
            <span className="text-xs text-slate-500 font-medium">Presets:</span>
            <button
              onClick={() => {
                setSteps(PRESET_SAAS);
                runCalculation(PRESET_SAAS);
              }}
              className="text-xs px-2.5 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 transition-colors"
            >
              SaaS Onboarding
            </button>
            <button
              onClick={() => {
                setSteps(PRESET_DOCS);
                runCalculation(PRESET_DOCS);
              }}
              className="text-xs px-2.5 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 transition-colors"
            >
              Docs to Pricing
            </button>
          </div>
        </div>

        {/* Step Editor list */}
        <div className="mt-6 space-y-2.5">
          <div className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
            Funnel Journey Steps
          </div>
          {steps.map((step, idx) => (
            <div
              key={idx}
              className="flex items-center gap-2 bg-slate-950/60 border border-slate-800/80 rounded-lg p-2.5 text-xs"
            >
              <span className="w-6 h-6 flex items-center justify-center rounded-md bg-blue-500/10 text-blue-400 font-bold text-[11px]">
                {idx + 1}
              </span>

              <select
                value={step.type}
                onChange={(e) =>
                  updateStep(idx, {
                    type: e.target.value as 'pageview' | 'event',
                    path: e.target.value === 'pageview' ? step.path || '/' : '',
                    name: e.target.value === 'event' ? step.name || 'click_button' : '',
                  })
                }
                className="bg-slate-900 border border-slate-700 text-slate-200 rounded px-2 py-1 outline-none font-medium"
              >
                <option value="pageview">Pageview</option>
                <option value="event">Custom Event</option>
              </select>

              {step.type === 'pageview' ? (
                <input
                  type="text"
                  placeholder="/pricing"
                  value={step.path || ''}
                  onChange={(e) => updateStep(idx, { path: e.target.value })}
                  className="flex-1 bg-slate-900 border border-slate-700 text-slate-200 rounded px-2.5 py-1 font-mono outline-none focus:border-blue-500"
                />
              ) : (
                <input
                  type="text"
                  placeholder="click_signup"
                  value={step.name || ''}
                  onChange={(e) => updateStep(idx, { name: e.target.value })}
                  className="flex-1 bg-slate-900 border border-slate-700 text-slate-200 rounded px-2.5 py-1 font-mono outline-none focus:border-blue-500"
                />
              )}

              {steps.length > 2 && (
                <button
                  onClick={() => removeStep(idx)}
                  className="p-1.5 text-slate-500 hover:text-red-400 transition-colors"
                  title="Remove step"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>
              )}
            </div>
          ))}

          <div className="flex items-center gap-3 pt-2">
            <button
              onClick={addStep}
              className="flex items-center gap-1.5 text-xs text-slate-300 hover:text-white bg-slate-800 hover:bg-slate-700 px-3 py-1.5 rounded-lg border border-slate-700 transition-colors"
            >
              <Plus className="w-3.5 h-3.5" />
              Add Step
            </button>
            <button
              onClick={() => runCalculation(steps)}
              disabled={loading}
              className="flex items-center gap-1.5 text-xs font-semibold text-white bg-blue-600 hover:bg-blue-500 px-4 py-1.5 rounded-lg transition-colors shadow-sm disabled:opacity-50"
            >
              <Play className="w-3.5 h-3.5 fill-current" />
              {loading ? 'Computing...' : 'Recalculate Funnel'}
            </button>
          </div>
        </div>
      </div>

      {error && (
        <div className="bg-red-500/10 border border-red-500/20 text-red-400 p-4 rounded-xl text-xs">
          {error}
        </div>
      )}

      {/* Funnel Visualizer */}
      {result && (
        <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-6">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-6 mb-6 border-b border-slate-800 gap-4">
            <div>
              <div className="text-xs text-slate-400">Total Journey Outcome</div>
              <div className="text-2xl font-bold text-white mt-1 flex items-baseline gap-2">
                <span>{result.overall_conversion_rate.toFixed(1)}%</span>
                <span className="text-xs font-medium text-slate-400">
                  ({result.total_completed} of {result.total_started} completed all steps)
                </span>
              </div>
            </div>
            <div className="flex items-center space-x-2 text-xs bg-blue-500/10 border border-blue-500/20 px-3 py-1.5 rounded-lg text-blue-400">
              <Sparkles className="w-4 h-4" />
              <span>Timeframe: {timeRange}</span>
            </div>
          </div>

          {/* Funnel Steps Visualization */}
          <div className="space-y-6">
            {result.steps.map((st, idx) => {
              const maxVisitors = result.steps[0].visitors || 1;
              const widthPct = Math.max(8, (st.visitors / maxVisitors) * 100);

              return (
                <div key={idx} className="relative">
                  <div className="flex items-center justify-between text-xs mb-1.5 font-medium">
                    <div className="flex items-center space-x-2">
                      <span className="w-5 h-5 rounded-full bg-slate-800 text-slate-300 flex items-center justify-center font-bold text-[10px]">
                        {idx + 1}
                      </span>
                      <span className="font-mono text-slate-200">{st.name}</span>
                      <span className="text-[10px] text-slate-500 uppercase px-1.5 py-0.2 rounded bg-slate-800/80">
                        {st.type}
                      </span>
                    </div>
                    <div className="flex items-center space-x-4">
                      <span className="text-white font-semibold">{st.visitors} visitors</span>
                      {idx > 0 && (
                        <span className="text-emerald-400 font-semibold">
                          {st.conversion_rate.toFixed(1)}% conv
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Funnel Bar */}
                  <div className="w-full bg-slate-950/80 rounded-lg h-9 overflow-hidden p-1 border border-slate-800/80 flex items-center">
                    <div
                      className={`h-full rounded-md transition-all duration-500 flex items-center px-3 text-xs font-medium ${
                        idx === 0
                          ? 'bg-gradient-to-r from-blue-600 to-blue-500 text-white'
                          : idx === result.steps.length - 1
                          ? 'bg-gradient-to-r from-emerald-600 to-emerald-500 text-white'
                          : 'bg-gradient-to-r from-indigo-600 to-blue-600 text-white'
                      }`}
                      style={{ width: `${widthPct}%` }}
                    >
                      <span className="truncate">{((st.visitors / maxVisitors) * 100).toFixed(1)}% of total</span>
                    </div>
                  </div>

                  {/* Drop-off connector */}
                  {idx < result.steps.length - 1 && (
                    <div className="py-2 pl-6 flex items-center space-x-3 text-xs text-rose-400/90 font-medium">
                      <div className="flex items-center space-x-1 bg-rose-500/10 px-2 py-0.5 rounded border border-rose-500/20">
                        <ArrowDown className="w-3 h-3" />
                        <span>{result.steps[idx + 1].drop_off_rate.toFixed(1)}% drop-off</span>
                      </div>
                      <span className="text-[11px] text-slate-500">
                        ({st.visitors - result.steps[idx + 1].visitors} users abandoned)
                      </span>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};
