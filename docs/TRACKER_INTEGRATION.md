# MicroHog Tracker Integration

MicroHog provides two client script options:
1. `microhog.js`: The core 2 KB analytics tracker for pageviews, clicks, and Core Web Vitals.
2. `microhog-recorder.js`: The DOM session replay recorder that streams anonymized mutation events using rrweb.

## Basic Setup

Add the script tag to the `<head>` of your website or web application:

```html
<script 
  defer 
  src="https://your-microhog-domain.com/microhog.js" 
  data-site="my-app"
  data-host="https://your-microhog-domain.com">
</script>
```

### Script Attributes

| Attribute | Required | Default | Description |
|---|---|---|---|
| `data-site` | No | `default` | The site identifier for multi-project instances. |
| `data-host` | No | Current script domain | Ingestion endpoint URL if serving script from a CDN. |
| `data-record` | No | `false` | Set to `true` to enable DOM session recording. |
| `data-auto-pageview` | No | `true` | Automatically sends `$pageview` on load and URL change. |

---

## Session Replay Recording

To enable session replay, add `data-record="true"`:

```html
<script 
  defer 
  src="https://your-microhog-domain.com/microhog.js" 
  data-site="my-app" 
  data-record="true">
</script>
```

### Privacy and Masking

The session recorder automatically masks sensitive data before transmission:
- `<input type="password">` fields are replaced with asterisks.
- `<input autocomplete="cc-number">` and card related inputs are masked.
- Elements with the CSS class `.microhog-mask` will have all inner text replaced.

Example:
```html
<div class="account-balance microhog-mask">
  $14,920.00
</div>
```

---

## Custom Event Tracking

MicroHog exposes a lightweight global API on `window.microhog`.

### Capturing Custom Events

```javascript
// Basic event
window.microhog.capture("newsletter_subscribed");

// Event with custom properties
window.microhog.capture("checkout_completed", {
  plan: "founder_annual",
  value: 99.00,
  currency: "USD"
});
```

### Declarative HTML Tracking

You can track clicks directly from markup without writing JavaScript:

```html
<button 
  data-microhog-event="clicked_cta_button"
  data-microhog-prop-tier="starter">
  Get Started
</button>
```

---

## Single Page Applications (SPA)

The tracker listens to browser navigation events (`popstate`, `pushState`, and `replaceState`). When a client-side route changes in Next.js, React Router, Vue, or Svelte, a `$pageview` event is dispatched automatically.
