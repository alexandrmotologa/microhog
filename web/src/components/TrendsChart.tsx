import React, { useState } from 'react';
import { TimeseriesPoint } from '../types/api';

interface TrendsChartProps {
  data: TimeseriesPoint[];
}

export const TrendsChart: React.FC<TrendsChartProps> = ({ data }) => {
  const [hoverIndex, setHoverIndex] = useState<number | null>(null);

  if (!data || data.length === 0) {
    return (
      <div className="h-64 flex items-center justify-center text-slate-500 text-sm">
        No traffic recorded for this timeframe
      </div>
    );
  }

  const width = 800;
  const height = 260;
  const padding = { top: 20, right: 20, bottom: 40, left: 40 };

  const chartWidth = width - padding.left - padding.right;
  const chartHeight = height - padding.top - padding.bottom;

  const maxViews = Math.max(...data.map((d) => d.views), 10);
  const maxVal = Math.ceil(maxViews * 1.15);

  const getX = (index: number) => {
    if (data.length <= 1) return padding.left + chartWidth / 2;
    return padding.left + (index / (data.length - 1)) * chartWidth;
  };

  const getY = (val: number) => {
    return padding.top + chartHeight - (val / maxVal) * chartHeight;
  };

  // Generate SVG path for views
  const viewsPoints = data.map((d, i) => `${getX(i)},${getY(d.views)}`).join(' ');
  const viewsArea = `${getX(0)},${padding.top + chartHeight} ${viewsPoints} ${getX(data.length - 1)},${padding.top + chartHeight}`;

  // Generate SVG path for visitors
  const visitorsPoints = data.map((d, i) => `${getX(i)},${getY(d.visitors)}`).join(' ');

  const activePoint = hoverIndex !== null ? data[hoverIndex] : null;

  return (
    <div className="relative w-full">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center space-x-5 text-xs">
          <div className="flex items-center space-x-2">
            <span className="w-3 h-3 rounded-full bg-blue-500"></span>
            <span className="text-slate-300 font-medium">Pageviews</span>
          </div>
          <div className="flex items-center space-x-2">
            <span className="w-3 h-3 rounded-full bg-cyan-400"></span>
            <span className="text-slate-300 font-medium">Unique Visitors</span>
          </div>
        </div>
        {activePoint && (
          <div className="text-xs bg-slate-800/90 border border-slate-700 rounded-lg px-3 py-1 flex items-center space-x-3">
            <span className="text-slate-400">{activePoint.timestamp}</span>
            <span className="text-blue-400 font-medium">{activePoint.views} views</span>
            <span className="text-cyan-400 font-medium">{activePoint.visitors} visitors</span>
          </div>
        )}
      </div>

      <div className="overflow-hidden">
        <svg
          viewBox={`0 0 ${width} ${height}`}
          className="w-full h-auto overflow-visible cursor-crosshair"
          onMouseLeave={() => setHoverIndex(null)}
        >
          <defs>
            <linearGradient id="viewsGrad" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#3b82f6" stopOpacity="0.35" />
              <stop offset="100%" stopColor="#3b82f6" stopOpacity="0.0" />
            </linearGradient>
          </defs>

          {/* Grid lines */}
          {[0, 0.25, 0.5, 0.75, 1].map((pct, idx) => {
            const y = padding.top + chartHeight * (1 - pct);
            const val = Math.round(maxVal * pct);
            return (
              <g key={idx}>
                <line
                  x1={padding.left}
                  y1={y}
                  x2={padding.left + chartWidth}
                  y2={y}
                  stroke="#1e293b"
                  strokeDasharray="4 4"
                />
                <text x={padding.left - 8} y={y + 3} textAnchor="end" fill="#64748b" fontSize="10">
                  {val}
                </text>
              </g>
            );
          })}

          {/* Area fill */}
          <polygon points={viewsArea} fill="url(#viewsGrad)" />

          {/* Lines */}
          <polyline
            fill="none"
            stroke="#3b82f6"
            strokeWidth="2.5"
            strokeLinecap="round"
            strokeLinejoin="round"
            points={viewsPoints}
          />
          <polyline
            fill="none"
            stroke="#06b6d4"
            strokeWidth="2"
            strokeDasharray="4 3"
            strokeLinecap="round"
            strokeLinejoin="round"
            points={visitorsPoints}
          />

          {/* X axis labels */}
          {data.map((d, i) => {
            // Show every few labels to prevent overlap
            const step = Math.ceil(data.length / 7);
            if (i % step !== 0 && i !== data.length - 1) return null;
            return (
              <text
                key={i}
                x={getX(i)}
                y={height - 12}
                textAnchor="middle"
                fill="#64748b"
                fontSize="10"
              >
                {d.timestamp.length > 10 ? d.timestamp.slice(11, 16) : d.timestamp.slice(5)}
              </text>
            );
          })}

          {/* Hover interactive vertical line and points */}
          {hoverIndex !== null && (
            <g>
              <line
                x1={getX(hoverIndex)}
                y1={padding.top}
                x2={getX(hoverIndex)}
                y2={padding.top + chartHeight}
                stroke="#94a3b8"
                strokeWidth="1.5"
                strokeDasharray="3 3"
              />
              <circle
                cx={getX(hoverIndex)}
                cy={getY(data[hoverIndex].views)}
                r="4.5"
                fill="#3b82f6"
                stroke="#0f172a"
                strokeWidth="2"
              />
              <circle
                cx={getX(hoverIndex)}
                cy={getY(data[hoverIndex].visitors)}
                r="4.5"
                fill="#06b6d4"
                stroke="#0f172a"
                strokeWidth="2"
              />
            </g>
          )}

          {/* Transparent interactive hover overlays */}
          {data.map((_, i) => {
            const stepWidth = chartWidth / data.length;
            const x = getX(i) - stepWidth / 2;
            return (
              <rect
                key={i}
                x={x}
                y={padding.top}
                width={stepWidth}
                height={chartHeight}
                fill="transparent"
                onMouseEnter={() => setHoverIndex(i)}
              />
            );
          })}
        </svg>
      </div>
    </div>
  );
};
