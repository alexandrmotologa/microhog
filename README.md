# MicroHog

Self-hosted product analytics, conversion funnels, and DOM session replay engine running in a single Go binary with SQLite storage.

MicroHog provides an alternative to complex, multi-service analytics systems. It requires under 50 MB of memory on a low-cost virtual private server. Ingestion operates via a 2 KB tracking script that works without persistent cookies, eliminating cookie consent banner requirements. The engine also records compressed DOM session replays with client-side privacy masking and computes conversion funnels.

---

## Key Features

- **Single Binary**: Self-contained Go executable with embedded SQLite storage and web user interface. No external dependencies like Kafka, Redis, or ClickHouse.
- **Cookieless Tracking**: Daily rotating cryptographic salts generate anonymous visitor identifiers. Meets GDPR, CCPA, and PECR standards without consent banners.
- **Lightweight Tracker**: Core analytics script is under 2 KB uncompressed. Captures pageviews, clicks, rage click frustration events, and Core Web Vitals.
- **Session Replay**: Records DOM mutations using rrweb with automatic masking of passwords, credit cards, and specified sensitive elements. Chunks are compressed using Zstandard.
- **Conversion Funnels & Persistence**: Computes drop-off rates and completion percentages across sequential user journeys with saved funnel persistence.
- **Frustration & Rage Click Detection**: Detects rapid consecutive clicks on dead or broken elements and flags sessions for quick replay review.
- **Core Web Vitals Monitoring**: Ingests LCP, INP, and CLS performance metrics with p75 rating breakdowns (Good, Needs Improvement, Poor).
- **Event Explorer & JSON Inspector**: Real-time event search and inspection modal for debugging custom event properties.
- **Multi-Site Management**: Create, switch, and manage multiple tracking tokens and domains directly within the dashboard.
- **Data Export & Portability**: Direct CSV and JSON export endpoints for raw events and session metadata.
- **Webhook Alerting**: Dispatches HTTP POST notifications to Slack, Discord, or custom endpoints on rage clicks and conversion milestones.
- **Real-Time Feed**: Live active visitor counts and Server-Sent Events stream for instant inspection.

---

## Quick Start

### 1. Download or Build

With Go installed:

```bash
git clone https://github.com/alexandrmotologa/microhog.git
cd microhog
go build -o microhog cmd/microhog/main.go
```

### 2. Start the Server

```bash
./microhog serve --port 8080 --db ./microhog.db
```

The web dashboard is accessible at `http://localhost:8080`.

### 3. Add the Tracking Snippet

Place this snippet into the `<head>` of your website:

```html
<script 
  defer 
  src="http://localhost:8080/microhog.js" 
  data-site="default"
  data-host="http://localhost:8080">
</script>
```

To enable session replay recording, include `data-record="true"`:

```html
<script 
  defer 
  src="http://localhost:8080/microhog.js" 
  data-site="default"
  data-record="true">
</script>
```

---

## Architecture Overview

```
Client Browser (microhog.js 2KB)
       │
       │ HTTP POST /e (Batching & Beacon API)
       ▼
Go Server (:8080)
       │
       ├─► Daily Salt Anonymizer (Zero IP / Cookie persistence)
       │
       ├─► Memory Ring Buffer (Flushes every 500ms or 200 events)
       │
       ▼
SQLite Database (WAL Mode + Zstandard Replay Storage)
       │
       ▼
Analytics Engine (Funnels, Aggregate Metrics, Realtime SSE)
       │
       ▼
Embedded Web Studio & Replay Player (React 19)
```

---

## Configuration Flags

| Flag | Default | Description |
|---|---|---|
| `--port` | `8080` | Port for HTTP server to listen on. |
| `--db` | `./microhog.db` | File path to SQLite database. |
| `--salt-secret` | Auto-generated | Secret used for daily rotating visitor hash. |
| `--retention-days` | `90` | Days to keep events and session recordings before automatic cleanup. |

---

## Documentation

- [System Architecture](docs/ARCHITECTURE.md)
- [REST API Reference](docs/API_REFERENCE.md)
- [Tracker Integration Guide](docs/TRACKER_INTEGRATION.md)
- [Production Deployment](docs/DEPLOYMENT.md)

---

## License

MIT License. See [LICENSE](LICENSE) for details.
