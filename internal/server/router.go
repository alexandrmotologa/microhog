package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/alexandrmotologa/microhog/internal/analytics"
	"github.com/alexandrmotologa/microhog/internal/ingest"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// RouterConfig bundles the handlers required to mount the HTTP router.
type RouterConfig struct {
	Collector       *ingest.Collector
	APIHandler      *APIHandler
	RealtimeTracker *analytics.RealtimeTracker
	StaticFileFS    http.FileSystem
}

// NewRouter sets up the HTTP router with middleware and endpoints.
func NewRouter(cfg *RouterConfig) http.Handler {
	mux := http.NewServeMux()

	// Ingestion endpoints
	mux.HandleFunc("POST /e", cfg.Collector.HandleEvent)
	mux.HandleFunc("POST /api/replay/chunk", cfg.Collector.HandleReplayChunk)

	// Analytics & Metrics
	mux.HandleFunc("GET /health", cfg.APIHandler.HandleHealth)
	mux.HandleFunc("GET /api/metrics", cfg.APIHandler.HandleMetrics)
	mux.HandleFunc("GET /api/realtime", cfg.APIHandler.HandleRealtime)
	mux.HandleFunc("GET /api/realtime/stream", cfg.RealtimeTracker.HandleSSE)
	mux.HandleFunc("POST /api/funnels", cfg.APIHandler.HandleFunnels)
	mux.HandleFunc("GET /api/funnels/saved", cfg.APIHandler.HandleSavedFunnels)
	mux.HandleFunc("POST /api/funnels/saved", cfg.APIHandler.HandleSavedFunnels)
	mux.HandleFunc("DELETE /api/funnels/saved/{id}", cfg.APIHandler.HandleDeleteSavedFunnel)

	// Core Web Vitals & Event Explorer
	mux.HandleFunc("GET /api/vitals", cfg.APIHandler.HandleWebVitals)
	mux.HandleFunc("GET /api/events/search", cfg.APIHandler.HandleEventSearch)

	// Sessions & Replays
	mux.HandleFunc("GET /api/sessions", cfg.APIHandler.HandleSessions)
	mux.HandleFunc("GET /api/sessions/{id}/replay", cfg.APIHandler.HandleSessionReplay)

	// Data Exports (CSV & JSON)
	mux.HandleFunc("GET /api/export/events", cfg.APIHandler.HandleExportEvents)
	mux.HandleFunc("GET /api/export/sessions", cfg.APIHandler.HandleExportSessions)

	// Webhooks
	mux.HandleFunc("GET /api/webhooks", cfg.APIHandler.HandleWebhooks)
	mux.HandleFunc("POST /api/webhooks", cfg.APIHandler.HandleWebhooks)
	mux.HandleFunc("DELETE /api/webhooks/{id}", cfg.APIHandler.HandleDeleteWebhook)
	mux.HandleFunc("POST /api/webhooks/test", cfg.APIHandler.HandleTestWebhook)

	// Sites Management
	mux.HandleFunc("GET /api/sites", cfg.APIHandler.HandleSites)
	mux.HandleFunc("POST /api/sites", cfg.APIHandler.HandleSites)
	mux.HandleFunc("DELETE /api/sites/{id}", cfg.APIHandler.HandleDeleteSite)

	// Database Seeding
	mux.HandleFunc("POST /api/seed", cfg.APIHandler.HandleSeed)

	// Static & Embedded dashboard handler
	if cfg.StaticFileFS != nil {
		fileServer := http.FileServer(cfg.StaticFileFS)
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			// Serve tracker scripts directly with cache headers
			if path == "/microhog.js" || path == "/microhog-recorder.js" {
				w.Header().Set("Cache-Control", "public, max-age=3600")
				fileServer.ServeHTTP(w, r)
				return
			}

			// Check if file exists in filesystem
			f, err := cfg.StaticFileFS.Open(strings.TrimPrefix(path, "/"))
			if err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}

			// SPA fallback: serve index.html for unrecognized routes (except /api/ or /e)
			if !strings.HasPrefix(path, "/api/") && path != "/e" {
				r.URL.Path = "/"
				fileServer.ServeHTTP(w, r)
				return
			}

			http.NotFound(w, r)
		})
	}

	return corsMiddleware(gzipMiddleware(mux))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Origin, Accept")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Avoid gzipping SSE streams
		if strings.Contains(r.Header.Get("Accept"), "text/event-stream") || r.URL.Path == "/api/realtime/stream" {
			next.ServeHTTP(w, r)
			return
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzw, r)
	})
}
