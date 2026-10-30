// Lightweight DOM Session Recorder conforming to rrweb event format
// Includes client-side PII masking for passwords, cards, and .microhog-mask elements

(function () {
  if (typeof window === 'undefined') return;
  if ((window as any).__microhog_recorder_active) return;
  (window as any).__microhog_recorder_active = true;

  const currentScript =
    document.currentScript ||
    document.querySelector('script[src*="microhog-recorder.js"]') ||
    document.querySelector('script[data-record]');

  const siteId = currentScript?.getAttribute('data-site') || 'default';
  const sessionId =
    currentScript?.getAttribute('data-session') ||
    (window as any).microhog?.getSessionId?.() ||
    sessionStorage.getItem('__mh_sid') ||
    'sess_' + Math.random().toString(36).substring(2, 10);

  let host = currentScript?.getAttribute('data-host') || '';
  if (!host && currentScript && (currentScript as HTMLScriptElement).src) {
    try {
      host = new URL((currentScript as HTMLScriptElement).src).origin;
    } catch {}
  }
  if (!host) {
    host = window.location.origin;
  }

  const endpoint = host.replace(/\/$/, '') + '/api/replay/chunk';

  let chunkIndex = 0;
  let eventBuffer: any[] = [];
  let nextNodeId = 1;
  const nodeMap = new WeakMap<Node, number>();

  function getNodeId(node: Node): number {
    let id = nodeMap.get(node);
    if (!id) {
      id = nextNodeId++;
      nodeMap.set(node, id);
    }
    return id;
  }

  // Event Types: 0: DomContentLoaded, 1: Load, 2: FullSnapshot, 3: IncrementalSnapshot
  function emitEvent(type: number, data: any) {
    const ev = {
      type: type,
      data: data,
      timestamp: Date.now(),
    };
    eventBuffer.push(ev);

    if (eventBuffer.length >= 40) {
      flushChunk();
    }
  }

  function flushChunk() {
    if (eventBuffer.length === 0) return;

    const payload = {
      session_id: sessionId,
      chunk_index: chunkIndex++,
      event_count: eventBuffer.length,
      events: eventBuffer,
    };
    eventBuffer = [];

    const body = JSON.stringify(payload);
    if (navigator.sendBeacon) {
      const blob = new Blob([body], { type: 'application/json' });
      if (navigator.sendBeacon(endpoint, blob)) return;
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

  // Periodic flush every 4 seconds
  setInterval(flushChunk, 4000);
  window.addEventListener('pagehide', flushChunk);
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'hidden') flushChunk();
  });

  // Serialization helper with automatic PII masking
  function serializeNode(node: Node): any {
    const id = getNodeId(node);

    if (node.nodeType === Node.DOCUMENT_TYPE_NODE) {
      const dt = node as DocumentType;
      return {
        id: id,
        type: 1,
        name: dt.name,
        publicId: dt.publicId || '',
        systemId: dt.systemId || '',
      };
    }

    if (node.nodeType === Node.TEXT_NODE) {
      let text = node.textContent || '';
      const parent = node.parentElement;
      if (parent && (parent.classList.contains('microhog-mask') || parent.tagName === 'INPUT')) {
        text = '••••••••';
      }
      return {
        id: id,
        type: 3,
        textContent: text,
      };
    }

    if (node.nodeType === Node.ELEMENT_NODE) {
      const el = node as HTMLElement;
      const tagName = el.tagName.toLowerCase();
      if (tagName === 'script' || tagName === 'noscript') {
        return null;
      }

      const attributes: Record<string, string> = {};
      for (const attr of Array.from(el.attributes)) {
        if (tagName === 'input' && (attr.name === 'value' || attr.name === 'placeholder')) {
          const inputType = el.getAttribute('type') || 'text';
          const autoComp = el.getAttribute('autocomplete') || '';
          if (
            inputType === 'password' ||
            inputType === 'email' ||
            autoComp.includes('cc-') ||
            el.classList.contains('microhog-mask')
          ) {
            attributes[attr.name] = '••••••••';
            continue;
          }
        }
        attributes[attr.name] = attr.value;
      }

      const childNodes: any[] = [];
      for (let i = 0; i < el.childNodes.length; i++) {
        const child = serializeNode(el.childNodes[i]);
        if (child) {
          childNodes.push(child);
        }
      }

      return {
        id: id,
        type: 2,
        tagName: tagName,
        attributes: attributes,
        childNodes: childNodes,
      };
    }

    return null;
  }

  function takeFullSnapshot() {
    const docNode: any = {
      id: getNodeId(document),
      type: 0,
      childNodes: [],
    };

    for (let i = 0; i < document.childNodes.length; i++) {
      const s = serializeNode(document.childNodes[i]);
      if (s) docNode.childNodes.push(s);
    }

    emitEvent(2, {
      node: docNode,
      initialOffset: {
        top: window.scrollY || document.documentElement.scrollTop,
        left: window.scrollX || document.documentElement.scrollLeft,
      },
    });
  }

  // 1. Initial Document Snapshot
  emitEvent(0, {}); // DomContentLoaded
  if (document.readyState === 'complete') {
    emitEvent(1, {}); // Load
    takeFullSnapshot();
  } else {
    window.addEventListener('load', () => {
      emitEvent(1, {});
      takeFullSnapshot();
    });
  }

  // 2. Mouse Move Tracking (Throttled to 60ms)
  let lastMove = 0;
  window.addEventListener(
    'mousemove',
    (e) => {
      const now = Date.now();
      if (now - lastMove < 60) return;
      lastMove = now;

      emitEvent(3, {
        source: 1, // MouseMove
        positions: [
          {
            x: e.clientX,
            y: e.clientY,
            id: getNodeId(e.target as Node),
            timeOffset: 0,
          },
        ],
      });
    },
    { passive: true }
  );

  // 3. Mouse Click Tracking
  window.addEventListener(
    'click',
    (e) => {
      emitEvent(3, {
        source: 2, // MouseInteraction
        type: 2, // Click
        id: getNodeId(e.target as Node),
        x: e.clientX,
        y: e.clientY,
      });
    },
    { passive: true }
  );

  // 4. Scroll Tracking (Throttled to 100ms)
  let lastScroll = 0;
  window.addEventListener(
    'scroll',
    () => {
      const now = Date.now();
      if (now - lastScroll < 100) return;
      lastScroll = now;

      emitEvent(3, {
        source: 3, // Scroll
        id: getNodeId(document),
        x: window.scrollX || document.documentElement.scrollLeft,
        y: window.scrollY || document.documentElement.scrollTop,
      });
    },
    { passive: true }
  );

  // 5. Window Resize Tracking
  window.addEventListener(
    'resize',
    () => {
      emitEvent(3, {
        source: 4, // ViewportResize
        width: window.innerWidth,
        height: window.innerHeight,
      });
    },
    { passive: true }
  );

  // 6. DOM Mutation Tracking using MutationObserver
  if (typeof MutationObserver !== 'undefined') {
    const observer = new MutationObserver((mutations) => {
      const texts: any[] = [];
      const adds: any[] = [];
      const removes: any[] = [];

      for (const m of mutations) {
        if (m.type === 'characterData' && m.target) {
          let val = m.target.textContent || '';
          if (m.target.parentElement?.classList.contains('microhog-mask')) {
            val = '••••••••';
          }
          texts.push({
            id: getNodeId(m.target),
            value: val,
          });
        } else if (m.type === 'childList') {
          for (const node of Array.from(m.addedNodes)) {
            const s = serializeNode(node);
            if (s) {
              adds.push({
                parentId: getNodeId(m.target),
                nextId: m.nextSibling ? getNodeId(m.nextSibling) : null,
                node: s,
              });
            }
          }
          for (const node of Array.from(m.removedNodes)) {
            removes.push({
              parentId: getNodeId(m.target),
              id: getNodeId(node),
            });
          }
        }
      }

      if (texts.length > 0 || adds.length > 0 || removes.length > 0) {
        emitEvent(3, {
          source: 0, // Mutation
          texts: texts,
          adds: adds,
          removes: removes,
        });
      }
    });

    observer.observe(document.documentElement, {
      childList: true,
      subtree: true,
      characterData: true,
    });
  }
})();
