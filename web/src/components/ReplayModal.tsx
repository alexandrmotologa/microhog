import React, { useState, useEffect, useRef } from 'react';
import { X, Play, Pause, RotateCcw, FastForward, ShieldAlert, Laptop } from 'lucide-react';

interface ReplayModalProps {
  sessionId: string;
  onClose: () => void;
}

export const ReplayModal: React.FC<ReplayModalProps> = ({ sessionId, onClose }) => {
  const [events, setEvents] = useState<any[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const [isPlaying, setIsPlaying] = useState<boolean>(true);
  const [currentTime, setCurrentTime] = useState<number>(0);
  const [duration, setDuration] = useState<number>(1000);
  const [speed, setSpeed] = useState<number>(1);
  const [cursorPos, setCursorPos] = useState<{ x: number; y: number; clicking: boolean }>({
    x: 100,
    y: 100,
    clicking: false,
  });

  const animRef = useRef<number | null>(null);
  const lastTickRef = useRef<number>(Date.now());
  const iframeRef = useRef<HTMLIFrameElement | null>(null);

  // 1. Fetch replay events
  useEffect(() => {
    async function loadReplay() {
      try {
        setLoading(true);
        setError(null);
        const res = await fetch(`/api/sessions/${encodeURIComponent(sessionId)}/replay`);
        if (!res.ok) throw new Error('No replay chunks recorded for this session');
        const data = await res.json();
        const evList = data.events || [];
        setEvents(evList);

        if (evList.length > 0) {
          const startTime = evList[0].timestamp || 0;
          const endTime = evList[evList.length - 1].timestamp || startTime + 1000;
          const totalDur = Math.max(1000, endTime - startTime);
          setDuration(totalDur);
        }
      } catch (err: any) {
        setError(err.message || 'Failed to load session recording');
      } finally {
        setLoading(false);
      }
    }
    loadReplay();
  }, [sessionId]);

  // 2. Render initial DOM snapshot into iframe
  useEffect(() => {
    if (!iframeRef.current || events.length === 0) return;

    // Find first FullSnapshot (type === 2)
    const snapshot = events.find((e) => e.type === 2);
    if (!snapshot || !snapshot.data?.node) return;

    const iframeDoc = iframeRef.current.contentDocument || iframeRef.current.contentWindow?.document;
    if (!iframeDoc) return;

    // Convert rrweb node tree to HTML string
    function renderNode(n: any): string {
      if (!n) return '';
      if (n.type === 3) {
        return n.textContent || '';
      }
      if (n.type === 2) {
        const tagName = n.tagName || 'div';
        let attrs = '';
        if (n.attributes) {
          for (const [k, v] of Object.entries(n.attributes)) {
            attrs += ` ${k}="${String(v).replace(/"/g, '&quot;')}"`;
          }
        }
        let children = '';
        if (n.childNodes) {
          for (const c of n.childNodes) {
            children += renderNode(c);
          }
        }
        return `<${tagName}${attrs}>${children}</${tagName}>`;
      }
      if (n.type === 0 && n.childNodes) {
        return n.childNodes.map(renderNode).join('');
      }
      return '';
    }

    const htmlContent = renderNode(snapshot.data.node);
    iframeDoc.open();
    iframeDoc.write(`
      <!DOCTYPE html>
      <html>
        <head>
          <meta charset="utf-8">
          <style>
            * { box-sizing: border-box; }
            body { pointer-events: none; margin: 0; user-select: none; }
          </style>
        </head>
        <body>
          ${htmlContent}
        </body>
      </html>
    `);
    iframeDoc.close();
  }, [events]);

  // 3. Playback clock loop
  useEffect(() => {
    if (!isPlaying || duration <= 0) return;

    lastTickRef.current = Date.now();

    const loop = () => {
      const now = Date.now();
      const delta = (now - lastTickRef.current) * speed;
      lastTickRef.current = now;

      setCurrentTime((prev) => {
        const next = prev + delta;
        if (next >= duration) {
          setIsPlaying(false);
          return duration;
        }

        // Update cursor position according to current timestamp
        if (events.length > 0) {
          const baseTimestamp = events[0].timestamp || 0;
          const targetTimestamp = baseTimestamp + next;

          // Find the closest mouse move before or at targetTimestamp
          for (let i = events.length - 1; i >= 0; i--) {
            const ev = events[i];
            if (ev.type === 3 && ev.timestamp <= targetTimestamp) {
              if (ev.data?.source === 1 && ev.data?.positions?.length > 0) {
                const pos = ev.data.positions[0];
                setCursorPos((c) => ({ ...c, x: pos.x, y: pos.y, clicking: false }));
                break;
              } else if (ev.data?.source === 2 && ev.data?.type === 2) {
                // Click
                setCursorPos((c) => ({ ...c, x: ev.data.x, y: ev.data.y, clicking: true }));
                break;
              }
            }
          }
        }

        return next;
      });

      animRef.current = requestAnimationFrame(loop);
    };

    animRef.current = requestAnimationFrame(loop);

    return () => {
      if (animRef.current) cancelAnimationFrame(animRef.current);
    };
  }, [isPlaying, duration, speed, events]);

  const toggleSpeed = () => {
    const speeds = [1, 2, 4];
    const nextIdx = (speeds.indexOf(speed) + 1) % speeds.length;
    setSpeed(speeds[nextIdx]);
  };

  const handleSeek = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = Number(e.target.value);
    setCurrentTime(val);
  };

  const formatMs = (ms: number) => {
    const totalSec = Math.floor(ms / 1000);
    const m = Math.floor(totalSec / 60);
    const s = totalSec % 60;
    return `${m}:${s < 10 ? '0' : ''}${s}`;
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 sm:p-6 animate-in fade-in duration-200">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-5xl h-[85vh] flex flex-col shadow-2xl overflow-hidden">
        {/* Top Bar */}
        <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-slate-950/60">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-lg bg-blue-500/10 text-blue-400 border border-blue-500/20">
              <Laptop className="w-4 h-4" />
            </div>
            <div>
              <div className="text-sm font-bold text-white flex items-center gap-2">
                Session Video Replay
                <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-800 text-slate-300">
                  {sessionId.slice(0, 14)}...
                </span>
              </div>
              <div className="text-xs text-slate-400 flex items-center gap-1.5 mt-0.5">
                <ShieldAlert className="w-3.5 h-3.5 text-emerald-400" />
                <span>Client-side PII masking active: passwords & cards obfuscated.</span>
              </div>
            </div>
          </div>

          <button
            onClick={onClose}
            className="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Viewport Area */}
        <div className="flex-1 relative bg-slate-950 flex items-center justify-center overflow-hidden">
          {loading && (
            <div className="flex flex-col items-center space-y-3 text-slate-400">
              <div className="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
              <span className="text-xs">Decompressing Zstandard replay stream...</span>
            </div>
          )}

          {error && (
            <div className="text-center p-8 max-w-md">
              <div className="text-rose-400 font-semibold text-sm mb-1">Replay Unavailable</div>
              <p className="text-xs text-slate-500">{error}</p>
            </div>
          )}

          {!loading && !error && (
            <div className="relative w-full h-full bg-slate-950 flex items-center justify-center">
              {/* Virtual Browser Window Frame */}
              <div className="w-full h-full relative overflow-hidden">
                <iframe
                  ref={iframeRef}
                  title="Session DOM"
                  className="w-full h-full border-none pointer-events-none"
                  sandbox="allow-same-origin"
                />

                {/* Simulated Mouse Cursor Overlay */}
                <div
                  className="absolute pointer-events-none transition-all duration-75 z-30"
                  style={{
                    transform: `translate(${cursorPos.x}px, ${cursorPos.y}px)`,
                    top: 0,
                    left: 0,
                  }}
                >
                  <svg
                    className="w-5 h-5 text-red-500 filter drop-shadow-md"
                    viewBox="0 0 24 24"
                    fill="currentColor"
                  >
                    <path d="M4 0l16 12-7 2-4 9z" />
                  </svg>
                  {cursorPos.clicking && (
                    <span className="absolute -top-2 -left-2 w-8 h-8 rounded-full border-2 border-red-400 animate-ping opacity-75"></span>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Player Bottom Control Bar */}
        <div className="px-6 py-4 bg-slate-950/90 border-t border-slate-800 flex flex-col gap-3">
          {/* Progress Timeline Scrubber */}
          <div className="flex items-center space-x-3 text-xs font-mono text-slate-400">
            <span>{formatMs(currentTime)}</span>
            <input
              type="range"
              min={0}
              max={duration}
              value={currentTime}
              onChange={handleSeek}
              className="flex-1 h-1.5 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-blue-500"
            />
            <span>{formatMs(duration)}</span>
          </div>

          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <button
                onClick={() => setIsPlaying(!isPlaying)}
                className="p-2.5 rounded-xl bg-blue-600 hover:bg-blue-500 text-white transition-colors shadow-sm"
              >
                {isPlaying ? <Pause className="w-4 h-4 fill-current" /> : <Play className="w-4 h-4 fill-current" />}
              </button>

              <button
                onClick={() => {
                  setCurrentTime(0);
                  setIsPlaying(true);
                }}
                className="p-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 transition-colors"
                title="Replay from start"
              >
                <RotateCcw className="w-4 h-4" />
              </button>

              <button
                onClick={toggleSpeed}
                className="flex items-center gap-1 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-medium font-mono transition-colors"
              >
                <FastForward className="w-3.5 h-3.5" />
                <span>{speed}x</span>
              </button>
            </div>

            <div className="text-xs text-slate-400 font-mono">
              Events: {events.length} frames
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
