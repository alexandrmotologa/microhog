import { observeVitals, WebVitals } from './vitals';

(function () {
  if (typeof window === 'undefined') return;

  // Prevent multiple initializations
  if ((window as any).__microhog_loaded) return;
  (window as any).__microhog_loaded = true;

  // Extract config from script tag attributes
  const currentScript =
    document.currentScript ||
    document.querySelector('script[data-site]') ||
    document.querySelector('script[src*="microhog.js"]');

  const siteId = currentScript?.getAttribute('data-site') || 'default';
  const autoPageview = currentScript?.getAttribute('data-auto-pageview') !== 'false';
  const enableRecord = currentScript?.getAttribute('data-record') === 'true';

  let host = currentScript?.getAttribute('data-host') || '';
  if (!host && currentScript && (currentScript as HTMLScriptElement).src) {
    try {
      const parsed = new URL((currentScript as HTMLScriptElement).src);
      host = parsed.origin;
    } catch {}
  }
  if (!host) {
    host = window.location.origin;
  }

  // Session ID maintained for current browser tab session (GDPR compliant, zero persistent cookie)
  const SESSION_KEY = '__mh_sid';
  let sessionId = '';
  try {
    sessionId = sessionStorage.getItem(SESSION_KEY) || '';
    if (!sessionId) {
      sessionId = 'mh_' + Math.random().toString(36).substring(2, 10) + Date.now().toString(36);
      sessionStorage.setItem(SESSION_KEY, sessionId);
    }
  } catch {
    sessionId = 'mh_' + Math.random().toString(36).substring(2, 10) + Date.now().toString(36);
  }

  // Event dispatcher
  function sendEvent(eventName: string, properties: Record<string, any> = {}) {
    const payload = {
      site_id: siteId,
      session_id: sessionId,
      event: eventName,
      url: window.location.href,
      path: window.location.pathname,
      referrer: document.referrer || '',
      properties: properties,
    };

    const endpoint = host.replace(/\/$/, '') + '/e';
    const body = JSON.stringify(payload);

    if (navigator.sendBeacon) {
      // Use Blob with text/plain to avoid preflight CORS options where supported
      const blob = new Blob([body], { type: 'text/plain' });
      const sent = navigator.sendBeacon(endpoint, blob);
      if (sent) return;
    }

    try {
      fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: body,
        keepalive: true,
      }).catch(() => {});
    } catch {}
  }

  // Track pageview
  let lastTrackedPath = '';
  function trackPageview() {
    const currentPath = window.location.pathname;
    lastTrackedPath = currentPath;
    sendEvent('$pageview', {
      title: document.title,
    });
  }

  // SPA navigation handling (History API & popstate)
  if (autoPageview) {
    trackPageview();

    const originalPushState = history.pushState;
    history.pushState = function (...args) {
      originalPushState.apply(this, args);
      if (window.location.pathname !== lastTrackedPath) {
        trackPageview();
      }
    };

    const originalReplaceState = history.replaceState;
    history.replaceState = function (...args) {
      originalReplaceState.apply(this, args);
      if (window.location.pathname !== lastTrackedPath) {
        trackPageview();
      }
    };

    window.addEventListener('popstate', () => {
      if (window.location.pathname !== lastTrackedPath) {
        trackPageview();
      }
    });
  }

  // Declarative click tracking & outbound link detection
  document.addEventListener(
    'click',
    (e) => {
      let target = e.target as HTMLElement | null;
      while (target && target !== document.body) {
        // 1. Declarative data-microhog-event attribute
        const customEvent = target.getAttribute('data-microhog-event');
        if (customEvent) {
          const props: Record<string, any> = {};
          for (const attr of Array.from(target.attributes)) {
            if (attr.name.startsWith('data-microhog-prop-')) {
              const propKey = attr.name.replace('data-microhog-prop-', '');
              props[propKey] = attr.value;
            }
          }
          sendEvent(customEvent, props);
          return;
        }

        // 2. Outbound link click
        if (target.tagName === 'A') {
          const href = (target as HTMLAnchorElement).href;
          if (href && href.startsWith('http') && !href.includes(window.location.hostname)) {
            sendEvent('$outbound_click', { href: href });
          }
          return;
        }

        target = target.parentElement;
      }
    },
    true
  );

  // Observe Web Vitals
  observeVitals((vitals: WebVitals) => {
    sendEvent('$web_vitals', vitals);
  });

  // Dynamic session recorder loader if data-record="true"
  if (enableRecord) {
    const recScript = document.createElement('script');
    recScript.defer = true;
    recScript.src = host.replace(/\/$/, '') + '/microhog-recorder.js';
    recScript.setAttribute('data-site', siteId);
    recScript.setAttribute('data-session', sessionId);
    recScript.setAttribute('data-host', host);
    document.head.appendChild(recScript);
  }

  // Public API
  (window as any).microhog = {
    capture: (event: string, properties?: Record<string, any>) => {
      sendEvent(event, properties || {});
    },
    getSessionId: () => sessionId,
    getSiteId: () => siteId,
  };
})();
