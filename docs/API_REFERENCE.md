# MicroHog API Reference

This document outlines the HTTP endpoints available in MicroHog for event ingestion, analytics queries, and studio administration.

## Ingestion Endpoints

### 1. Ingest Event
`POST /e`

Accepts an individual event or an array of event objects. Also accepts `text/plain` from `navigator.sendBeacon`.

**Payload Schema:**
```json
{
  "site_id": "default",
  "session_id": "mh_a1b2c3d4",
  "event": "$pageview",
  "url": "https://example.com/pricing",
  "path": "/pricing",
  "referrer": "https://news.ycombinator.com",
  "properties": {
    "plan": "pro"
  }
}
```

**Response:**
- `200 OK`: `{"status":"ok","received":1}`

---

### 2. Ingest Session Replay Chunk
`POST /api/replay/chunk`

Accepts a chunk of recorded DOM mutations associated with a session.

**Payload Schema:**
```json
{
  "session_id": "mh_a1b2c3d4",
  "chunk_index": 0,
  "events": [...]
}
```

---

## Analytics Endpoints

### 3. Overview Metrics
`GET /api/metrics?site_id=default&range=7d`

Returns aggregate counters and breakdown tables for referrers, browsers, operating systems, devices, and countries.

---

### 4. Real-Time Active Users & SSE Stream
- `GET /api/realtime?site_id=default`: Snapshot of active visitors in the last 5 minutes.
- `GET /api/realtime/stream?site_id=default`: Server-Sent Events stream delivering live telemetry.

---

### 5. Conversion Funnel Calculation
`POST /api/funnels`

Computes sequential conversion percentages and drop-offs.

**Payload Schema:**
```json
{
  "site_id": "default",
  "steps": [
    { "type": "pageview", "path": "/" },
    { "type": "pageview", "path": "/pricing" },
    { "type": "event", "name": "signup_completed" }
  ],
  "range": "7d"
}
```

---

### 6. Saved Funnels Persistence
- `GET /api/funnels/saved?site_id=default`: List saved funnels for a project.
- `POST /api/funnels/saved`: Persist a new named funnel.
- `DELETE /api/funnels/saved/{id}`: Delete a saved funnel.

---

### 7. Core Web Vitals Report
`GET /api/vitals?site_id=default&range=7d`

Returns 75th percentile (p75) and median values for Largest Contentful Paint (LCP), Interaction to Next Paint (INP), and Cumulative Layout Shift (CLS) alongside page-level breakdowns.

---

### 8. Event Explorer & Search
`GET /api/events/search?site_id=default&query=pricing&event_name=all&limit=50&offset=0`

Searches events by path, visitor ID, or custom JSON properties.

---

### 9. List User Sessions
`GET /api/sessions?site_id=default&limit=50&offset=0&rage_only=false`

Returns user sessions. Set `rage_only=true` to isolate sessions where rapid frustration clicks occurred.

---

### 10. Fetch Session Replay
`GET /api/sessions/{id}/replay`

Retrieves decompressed DOM mutation streams for playback inside the web player.

---

### 11. Data Exports
- `GET /api/export/events?site_id=default&range=7d&format=csv`: Download raw events as a CSV file.
- `GET /api/export/events?site_id=default&range=7d&format=json`: Download raw events as JSON.
- `GET /api/export/sessions?site_id=default&range=7d&format=csv`: Download sessions as a CSV file.
- `GET /api/export/sessions?site_id=default&range=7d&format=json`: Download sessions as JSON.

---

### 12. Outbound Webhooks (Discord, Slack, HTTP)
- `GET /api/webhooks?site_id=default`: List registered webhooks.
- `POST /api/webhooks`: Register a webhook URL and event trigger filter.
- `DELETE /api/webhooks/{id}`: Remove a registered webhook.
- `POST /api/webhooks/test`: Send a test notification ping to a target URL.

---

### 13. Multi-Site Management
- `GET /api/sites`: List all tracked websites and projects.
- `POST /api/sites`: Create a new project (`id`, `name`, `domain`).
- `DELETE /api/sites/{id}`: Remove a project (excluding `default`).
