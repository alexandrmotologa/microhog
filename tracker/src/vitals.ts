// Core Web Vitals collector using native PerformanceObserver
export interface WebVitals {
  lcp?: number;
  inp?: number;
  cls?: number;
}

export function observeVitals(onReport: (vitals: WebVitals) => void) {
  if (typeof window === 'undefined' || !('PerformanceObserver' in window)) {
    return;
  }

  const vitals: WebVitals = {};
  let clsValue = 0;

  try {
    // 1. Largest Contentful Paint (LCP)
    const lcpObserver = new PerformanceObserver((entryList) => {
      const entries = entryList.getEntries();
      const lastEntry = entries[entries.length - 1];
      if (lastEntry) {
        vitals.lcp = Math.round(lastEntry.startTime);
      }
    });
    lcpObserver.observe({ type: 'largest-contentful-paint', buffered: true });
  } catch {}

  try {
    // 2. Cumulative Layout Shift (CLS)
    const clsObserver = new PerformanceObserver((entryList) => {
      for (const entry of entryList.getEntries()) {
        if (!(entry as any).hadRecentInput) {
          clsValue += (entry as any).value;
          vitals.cls = Math.round(clsValue * 1000) / 1000;
        }
      }
    });
    clsObserver.observe({ type: 'layout-shift', buffered: true });
  } catch {}

  try {
    // 3. Interaction to Next Paint (INP)
    const inpObserver = new PerformanceObserver((entryList) => {
      for (const entry of entryList.getEntries()) {
        const duration = entry.duration;
        if (!vitals.inp || duration > vitals.inp) {
          vitals.inp = Math.round(duration);
        }
      }
    });
    inpObserver.observe({ type: 'first-input', buffered: true });
  } catch {}

  // Flush vitals when tab becomes hidden or unloads
  const flush = () => {
    if (vitals.lcp || vitals.cls || vitals.inp) {
      onReport({ ...vitals });
    }
  };

  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'hidden') {
      flush();
    }
  });
  window.addEventListener('pagehide', flush);
}
