# MicroHog: Master Implementation Plan

> Lightweight, single-binary self-hosted product analytics, conversion funnels, and DOM session replay engine.

MicroHog provides an alternative to complex, multi-service analytics platforms. It runs as a single Go binary on a low-cost virtual private server using under 50 MB of memory. Events are ingested through a 2 KB tracking script that works without persistent cookies, eliminating the requirement for cookie consent banners. The engine records compressed DOM mutation sessions with client-side privacy masking and computes multi-step conversion funnels.

---

## 1. System Architecture

```
+-----------------------------------------------------------------------------------+
|                                     MICROHOG                                      |
+-----------------------------------------+-----------------------------------------+
                                          |
        +---------------------------------+---------------------------------+
        |                                                                   |
+-------v-------------------------+                 +-----------------------v-------+
|    2KB Zero-Cookie Tracker      |                 |    High-Throughput Ingest     |
|  * Pageviews, Clicks, Events    |                 |  * Go HTTP Endpoint (:8080)   |
|  * Anonymous Daily Device Salt  |                 |  * In-Memory Ring Buffer      |
|  * Web Vitals (LCP, INP, CLS)   |                 |  * Batch Inserts to SQLite    |
+-------+-------------------------+                 +-----------------------+-------+
        |                                                                   |
        +---------------------------------+---------------------------------+
                                          |
                        +-----------------v-----------------+
                        |   DOM Session Replay Recorder     |
                        |  * Compressed Mutation Stream     |
                        |  * PII and Input Auto-Masking     |
                        |  * Coordinate Throttling          |
                        +-----------------+-----------------+
                                          |
                        +-----------------v-----------------+
                        |    Analytics & Funnel Engine      |
                        |  * Multi-Step Conversion Funnels  |
                        |  * Real-Time Active Users (SSE)   |
                        |  * Breakdown by OS, Geo, Device   |
                        +-----------------+-----------------+
                                          |
                        +-----------------v-----------------+
                        |   Embedded Studio & Replay Player |
                        |  * React 19 Single Page App       |
                        |  * Replay Video Player            |
                        |  * Single Binary Go Embed         |
                        +-----------------------------------+
```

---

## 2. Directory Structure

```
microhog/
├── cmd/
│   └── microhog/
│       └── main.go               # CLI entrypoint (serve, seed, prune)
├── internal/
│   ├── analytics/
│   │   ├── funnels.go            # Conversion funnel calculation engine
│   │   ├── metrics.go            # Aggregated statistics (pages, referrers, devices)
│   │   ├── realtime.go           # Active visitor tracker with SSE stream
│   │   └── seeder.go             # Synthetic data generator for demo and testing
│   ├── config/
│   │   └── config.go             # Environment and CLI flags configuration
│   ├── ingest/
│   │   ├── anonymizer.go         # Daily salt hashing without IP persistence
│   │   ├── buffer_pool.go        # Thread-safe ring buffer and batch flush
│   │   └── collector.go          # HTTP handler for event and replay ingestion
│   ├── server/
│   │   ├── api.go                # REST API handlers for studio dashboard
│   │   ├── router.go             # HTTP router with CORS and gzip compression
│   │   └── static.go             # Embedded static frontend and tracker assets
│   └── storage/
│       ├── event_store.go        # Event queries and batch insert statements
│       ├── replay_store.go       # Zstandard compressed replay storage
│       ├── retention.go          # Automated cleanup policy runner
│       └── sqlite_db.go          # SQLite database connection and WAL setup
├── tracker/
│   ├── src/
│   │   ├── recorder.ts           # rrweb DOM session recorder with PII masking
│   │   ├── tracker.ts            # 2KB core tracker script (pageviews, clicks, vitals)
│   │   └── vitals.ts             # Web Vitals observer (LCP, INP, CLS)
│   ├── esbuild.config.mjs        # Bundler configuration targeting sub-2KB output
│   ├── package.json
│   └── tsconfig.json
├── web/
│   ├── src/
│   │   ├── components/
│   │   │   ├── FunnelChart.tsx   # Visual multi-step conversion funnel
│   │   │   ├── MetricCard.tsx    # Key metric indicator cards
│   │   │   ├── RealtimeCard.tsx  # Active visitor count and live stream table
│   │   │   ├── ReplayModal.tsx   # rrweb replay player with scrubber
│   │   │   ├── SessionTable.tsx  # User session list with duration and metadata
│   │   │   ├── SetupGuide.tsx    # Tracking script embed instructions
│   │   │   └── TrendsChart.tsx   # Time-series traffic visualization
│   │   ├── hooks/
│   │   │   └── useAnalytics.ts   # Data fetching and live SSE stream hook
│   │   ├── types/
│   │   │   └── api.ts            # Type definitions for API models
│   │   ├── App.tsx               # Main application layout and view tabs
│   │   ├── index.css             # Tailwind styling
│   │   └── main.tsx              # React 19 entrypoint
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── docs/
│   ├── API_REFERENCE.md          # Full REST and Ingestion API specification
│   ├── ARCHITECTURE.md           # Architecture, data flow, and privacy details
│   ├── DEPLOYMENT.md             # Docker, systemd, and VPS setup instructions
│   └── TRACKER_INTEGRATION.md    # Script embedding, masking rules, custom events
├── go.mod
├── go.sum
├── Dockerfile
├── .gitignore
├── LICENSE
└── README.md
```

---

## 3. Implementation Steps

### Phase 1: High-Speed Event Ingest and Privacy Anonymizer
- [ ] Implement `internal/config/config.go` with port, database path, salt secret, and retention days.
- [ ] Implement `internal/ingest/anonymizer.go`:
  - Calculate `VisitorID = SHA256(IP + UserAgent + DailySalt + SiteID)`.
  - Discard raw IP addresses immediately after country detection.
- [ ] Implement `internal/ingest/buffer_pool.go`:
  - Thread-safe buffer capturing events in memory channels.
  - Periodic batch write trigger (every 500ms or 200 events).
- [ ] Implement `internal/ingest/collector.go`:
  - HTTP handlers for `POST /e` (single/batch events) and `POST /api/replay/chunk`.

### Phase 2: Embedded SQLite Storage Engine
- [ ] Implement `internal/storage/sqlite_db.go`:
  - SQLite configuration with WAL journal mode and normal synchronous flag.
  - Table definitions: `sites`, `events`, `sessions`, `replays`.
- [ ] Implement `internal/storage/event_store.go`:
  - Prepared batch insert statements.
  - Query methods for time ranges, filters, aggregated counters, and breakdown statistics.
- [ ] Implement `internal/storage/replay_store.go`:
  - Zstandard compressed chunk persistence and retrieval.
- [ ] Implement `internal/storage/retention.go`:
  - Periodic deletion of events and replay data exceeding configured retention period.

### Phase 3: Lightweight Tracker Script and DOM Session Recorder
- [ ] Implement `tracker/src/vitals.ts` for LCP, INP, and CLS performance metrics.
- [ ] Implement `tracker/src/tracker.ts`:
  - Zero-dependency client tracking script under 2 KB.
  - Automatic `$pageview` capture on initial load and SPA history changes (`popstate`, `pushState`).
  - Automatic outbound link click and custom click attributes (`data-microhog-event`).
  - Custom event dispatch via `window.microhog.capture(name, props)`.
  - Use `navigator.sendBeacon` with `fetch(..., { keepalive: true })` fallback.
- [ ] Implement `tracker/src/recorder.ts`:
  - DOM session recording stream.
  - Sensitive input masking (`password`, `credit card`, `.microhog-mask`).
  - Chunked transmission to `/api/replay/chunk`.
- [ ] Configure `tracker/esbuild.config.mjs` for minified compilation into `internal/server/assets/`.

### Phase 4: Analytics Engine, Funnels and Real-Time SSE
- [ ] Implement `internal/analytics/realtime.go`:
  - In-memory tracker for active visitors in the last 5 minutes.
  - Server-Sent Events (SSE) broadcaster at `GET /api/realtime/stream`.
- [ ] Implement `internal/analytics/metrics.go`:
  - Aggregations for total visitors, views, bounce rate, average session length.
  - Breakdowns for pages, referrers, countries, browsers, devices, and operating systems.
- [ ] Implement `internal/analytics/funnels.go`:
  - Sequential step matching for conversion funnels within configurable timeframes.
  - Drop-off rates, conversion percentages, and completion time calculations.
- [ ] Implement `internal/analytics/seeder.go`:
  - Realistic synthetic data generator for immediate testing and verification.

### Phase 5: Embedded Web Studio (React 19)
- [ ] Setup `web/` with Vite, React 19, TypeScript, and Tailwind CSS.
- [ ] Implement `RealtimeCard.tsx`: live visitor counter, active paths, and country activity.
- [ ] Implement `TrendsChart.tsx`: time-series visual for views and unique visitors.
- [ ] Implement `FunnelChart.tsx`: interactive funnel builder with step drop-off visualization.
- [ ] Implement `SessionTable.tsx` and `ReplayModal.tsx`: recorded sessions list and DOM replay player with playback controls.
- [ ] Implement `SetupGuide.tsx`: embed snippet generator with site key and recording toggles.

### Phase 6: Single-Binary Assembly, CLI and Documentation
- [ ] Implement `internal/server/static.go` with `//go:embed` for frontend and tracker files.
- [ ] Implement `cmd/microhog/main.go` with CLI subcommands:
  - `serve`: start web server and ingestion engine.
  - `seed`: generate realistic demo records into database.
  - `prune`: run retention cleanup manually.
- [ ] Add `Dockerfile` for container deployment.
- [ ] Add complete documentation in `docs/` and root `README.md`.
- [ ] Write integration and unit tests validating end-to-end event collection, funnel computation, and replay storage.
