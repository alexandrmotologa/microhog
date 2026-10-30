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
|  * Rage Click Frustration Signal|                 |  * Asynchronous Webhooks      |
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
                        |  * Saved Funnels Persistence      |
                        |  * Core Web Vitals (p75 Scores)   |
                        |  * Real-Time Active Users (SSE)   |
                        |  * Event Explorer & Search        |
                        |  * CSV / JSON Data Exporter       |
                        +-----------------+-----------------+
                                          |
                        +-----------------v-----------------+
                        |   Embedded Studio & Replay Player |
                        |  * React 19 Single Page App       |
                        |  * Replay Video Player            |
                        |  * Multi-Site Management UI       |
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
│   │   ├── seeder.go             # Synthetic data generator for demo and testing
│   │   └── webhook_dispatcher.go # Outbound webhook notifier (Discord, Slack, HTTP)
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
│       ├── event_store.go        # Event queries, vitals, search, and CSV export
│       ├── funnel_store.go       # Persistent saved funnels CRUD
│       ├── models.go             # Domain types and query structures
│       ├── replay_store.go       # Zstandard compressed replay storage
│       ├── retention.go          # Automated cleanup policy runner
│       ├── sqlite_db.go          # SQLite database connection and WAL setup
│       └── webhook_store.go      # Webhooks configuration CRUD
├── tracker/
│   ├── src/
│   │   ├── recorder.ts           # rrweb DOM session recorder with PII masking
│   │   ├── tracker.ts            # 2KB core tracker script (pageviews, clicks, rage clicks, vitals)
│   │   └── vitals.ts             # Web Vitals observer (LCP, INP, CLS)
│   ├── esbuild.config.mjs        # Bundler configuration targeting sub-2KB output
│   ├── package.json
│   └── tsconfig.json
├── web/
│   ├── src/
│   │   ├── components/
│   │   │   ├── EventExplorer.tsx # Raw event search and JSON property inspector
│   │   │   ├── FunnelChart.tsx   # Visual multi-step conversion funnel and saved funnels
│   │   │   ├── MetricCard.tsx    # Key metric indicator cards
│   │   │   ├── RealtimeCard.tsx  # Active visitor count and live stream table
│   │   │   ├── ReplayModal.tsx   # rrweb replay player with scrubber
│   │   │   ├── SessionTable.tsx  # User session list with rage click badge and filter
│   │   │   ├── SetupGuide.tsx    # Tracking script embed instructions and webhook manager
│   │   │   ├── SiteModal.tsx     # Multi-project and site switcher modal
│   │   │   ├── TrendsChart.tsx   # Time-series traffic visualization
│   │   │   └── WebVitalsTab.tsx  # Core Web Vitals p75 cards and route benchmarks
│   │   ├── hooks/
│   │   │   └── useAnalytics.ts   # Data fetching, SSE stream, and API hooks
│   │   ├── types/
│   │   │   └── api.ts            # Type definitions for API models
│   │   ├── App.tsx               # Main application layout, site switcher, and tabs
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

## 3. Implemented Capabilities

- [x] **High-Speed Event Ingest and Privacy Anonymizer**: Daily rotating salt visitor hashing, raw IP discarded immediately, thread-safe memory ring buffer with batch flushes to SQLite WAL.
- [x] **Rage Click & Frustration Signal Detection**: Automatic client-side detection of rapid clicks (<600ms, >=3 clicks), tagging sessions with `has_rage_click=true`, and rage click filtering in the session list.
- [x] **Core Web Vitals Reporting**: Automatic tracking of LCP, INP, and CLS via `vitals.ts` with 75th percentile (p75) calculations and Google benchmark ratings (Good, Needs Improvement, Poor).
- [x] **Conversion Funnels & Persistence**: Multi-step sequential conversion calculations, drop-off rates, and full saving/management of custom named funnels in SQLite.
- [x] **Event Explorer & Property Inspector**: Searchable event query engine with modal inspector for custom JSON telemetry payloads.
- [x] **Multi-Site Management**: Dedicated interface to create, track, and switch between separate projects and micro-SaaS domains.
- [x] **CSV & JSON Data Exporters**: Stream raw events and session metrics for spreadsheets and external backups.
- [x] **Outbound Webhooks**: Asynchronous notifications dispatched to Discord, Slack, or generic HTTP endpoints when conversions or rage clicks occur.
- [x] **DOM Session Replay & Player**: Zstandard compressed mutation stream storage with client-side masking of passwords, credit cards, and `.microhog-mask` elements. Interactive player modal with scrubber, speed controls, and simulated mouse trajectory.
- [x] **Single-Binary Assembly**: Embedded static assets and tracker scripts using `//go:embed all:dist`.
