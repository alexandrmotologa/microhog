# MicroHog Architecture

MicroHog operates as a single executable that combines an ingestion engine, SQLite storage, analytics aggregation, and an embedded web user interface.

## System Overview

```
                      +-----------------------------+
                      |       Client Browser        |
                      |  (microhog.js / 2 KB script) |
                      +--------------+--------------+
                                     |
                                     | HTTP POST /e
                                     v
                      +-----------------------------+
                      |       Go HTTP Server        |
                      |   - CORS & Gzip Handling    |
                      |   - Daily Salt Anonymizer   |
                      +--------------+--------------+
                                     |
                                     | In-memory channel
                                     v
                      +-----------------------------+
                      |       Buffered Ingestion    |
                      |   Flushes every 500ms / 200 |
                      +--------------+--------------+
                                     |
                                     | SQLite WAL Write
                                     v
                      +-----------------------------+
                      |       SQLite Database       |
                      |   - events, sessions        |
                      |   - replays (zstd chunks)   |
                      +--------------+--------------+
                                     ^
                                     | SQL queries
                      +--------------+--------------+
                      |       Analytics Engine      |
                      |   - Realtime sliding window |
                      |   - Conversion Funnels      |
                      |   - Aggregate Metrics       |
                      +--------------+--------------+
                                     |
                                     | JSON REST / SSE
                                     v
                      +-----------------------------+
                      |    Embedded React 19 UI     |
                      +-----------------------------+
```

## Privacy Model and Anonymization

MicroHog does not set tracking cookies in the browser. It computes an anonymous visitor identifier using a daily rotating cryptographic salt:

`VisitorID = SHA256(IP + UserAgent + DailySalt + SiteID)`

Properties of this approach:
1. Cross-day tracking is prevented because the salt rotates every 24 hours (UTC).
2. Raw IP addresses are discarded in memory immediately after country lookup. The database never stores raw IP addresses.
3. The tracking script meets GDPR, CCPA, and PECR privacy rules without requiring a cookie consent banner.

## Storage and Write Path

The storage layer uses SQLite configured with Write-Ahead Logging (WAL):

```sql
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA busy_timeout = 5000;
PRAGMA cache_size = -20000;
```

Incoming events pass through a Go channel buffer (`buffer_pool.go`). A background worker drains the channel and writes batches using a single database transaction. This prevents write lock contention and achieves over 5,000 ingested events per second on single-core virtual machines.

## Session Replay Compression

DOM session replays record mutations, mouse coordinates, and scroll positions using the `rrweb` protocol. Before transmitting data to the server:
- Text inputs of type `password`, `email`, and inputs with autocomplete attributes for credit cards are masked with asterisks.
- Elements with the CSS class `.microhog-mask` have their text content obfuscated.
- The server compresses incoming JSON mutation chunks with Zstandard (`zstd`) before saving to the `replays` table, reducing storage by roughly 85% compared to raw JSON.
