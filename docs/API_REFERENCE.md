# MicroHog API Reference

This document outlines the HTTP endpoints available in MicroHog for event ingestion and dashboard data access.

## Ingestion Endpoints

### 1. Ingest Event
`POST /e`

Accepts an individual event or an array of event objects.

**Headers:**
- `Content-Type: application/json`

**Payload Schema:**
```json
{
  "site_id": "default",
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
  "session_id": "a9b8c7d6-e5f4",
  "chunk_index": 0,
  "events_json": "[...]"
}
```

**Response:**
- `200 OK`: `{"status":"ok"}`

---

## Analytics Endpoints

### 3. Overview Metrics
`GET /api/metrics?site_id=default&range=7d`

Returns aggregate counters and breakdown tables.

**Query Parameters:**
- `site_id`: Site identifier (default: `default`)
- `range`: Time window (`24h`, `7d`, `30d`, `90d`)

**Response Schema:**
```json
{
  "summary": {
    "pageviews": 14250,
    "unique_visitors": 4820,
    "bounce_rate": 38.4,
    "avg_duration_seconds": 142
  },
  "timeseries": [
    { "timestamp": "2026-10-24T00:00:00Z", "views": 1940, "visitors": 680 }
  ],
  "top_pages": [
    { "path": "/", "views": 8200 },
    { "path": "/pricing", "views": 2100 }
  ],
  "top_referrers": [
    { "referrer": "direct", "count": 6200 },
    { "referrer": "google.com", "count": 3100 }
  ],
  "browsers": [
    { "name": "Chrome", "count": 8900 },
    { "name": "Safari", "count": 3400 }
  ],
  "countries": [
    { "code": "US", "count": 6200 },
    { "code": "DE", "count": 1800 }
  ],
  "devices": [
    { "type": "desktop", "count": 9800 },
    { "type": "mobile", "count": 4450 }
  ]
}
```

---

### 4. Real-Time Active Users
`GET /api/realtime?site_id=default`

Returns visitor counts and activity recorded in the past 5 minutes.

---

### 5. Real-Time Event Stream (SSE)
`GET /api/realtime/stream?site_id=default`

Establishes a Server-Sent Events stream emitting events as they arrive at the ingestion buffer.

**Event Format:**
```
data: {"event":"$pageview","path":"/checkout","country":"US","timestamp":"2026-10-30T11:20:00Z"}
```

---

### 6. Conversion Funnel Calculation
`POST /api/funnels`

Computes conversion percentages across an ordered list of steps.

**Payload Schema:**
```json
{
  "site_id": "default",
  "steps": [
    { "type": "pageview", "path": "/" },
    { "type": "pageview", "path": "/pricing" },
    { "type": "event", "name": "signup_completed" }
  ],
  "time_window_hours": 24
}
```

**Response Schema:**
```json
{
  "steps": [
    { "step_index": 0, "name": "/", "visitors": 1200, "conversion_rate": 100.0, "drop_off_rate": 0.0 },
    { "step_index": 1, "name": "/pricing", "visitors": 480, "conversion_rate": 40.0, "drop_off_rate": 60.0 },
    { "step_index": 2, "name": "signup_completed", "visitors": 120, "conversion_rate": 25.0, "drop_off_rate": 75.0 }
  ],
  "overall_conversion_rate": 10.0
}
```

---

### 7. List User Sessions
`GET /api/sessions?site_id=default&limit=50&offset=0`

Returns recent sessions with metadata and indicators for available replays.

---

### 8. Fetch Session Replay
`GET /api/sessions/{id}/replay`

Retrieves the decompressed DOM mutation stream for playback inside the web player.
