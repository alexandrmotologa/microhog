package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexandrmotologa/microhog/internal/analytics"
	"github.com/alexandrmotologa/microhog/internal/config"
	"github.com/alexandrmotologa/microhog/internal/ingest"
	"github.com/alexandrmotologa/microhog/internal/server"
	"github.com/alexandrmotologa/microhog/internal/storage"
)

var (
	Version   = "1.0.0"
	BuildDate = "2026-10-30"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version", "-v":
			fmt.Printf("MicroHog v%s (built %s)\n", Version, BuildDate)
			return
		case "seed":
			runSeed(os.Args[2:])
			return
		case "prune":
			runPrune(os.Args[2:])
			return
		case "serve":
			runServe(os.Args[2:])
			return
		}
	}

	// Default to serve with any passed flags
	runServe(os.Args[1:])
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "HTTP server listening port")
	dbPath := fs.String("db", "./microhog.db", "Path to SQLite database file")
	saltSecret := fs.String("salt-secret", "", "Secret key for daily salt anonymization")
	retentionDays := fs.Int("retention-days", 90, "Retention period in days for events and replays")
	autoSeed := fs.Bool("seed", false, "Seed demo data on startup if database is empty")
	_ = fs.Parse(args)

	cfg := config.Load()
	if *port != 8080 || cfg.Port == 0 {
		cfg.Port = *port
	}
	if *dbPath != "./microhog.db" || cfg.DBPath == "" {
		cfg.DBPath = *dbPath
	}
	if *saltSecret != "" {
		cfg.SaltSecret = *saltSecret
	}
	if *retentionDays != 90 {
		cfg.RetentionDays = *retentionDays
	}

	log.Printf("[MicroHog] Starting v%s on port %d (DB: %s)", Version, cfg.Port, cfg.DBPath)

	// 1. Initialize SQLite Database
	db, err := storage.OpenDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("[MicroHog] Failed to open database: %v", err)
	}
	defer db.Close()

	// 2. Initialize Storage Stores
	eventStore := storage.NewEventStore(db)
	replayStore, err := storage.NewReplayStore(db, eventStore)
	if err != nil {
		log.Fatalf("[MicroHog] Failed to initialize replay store: %v", err)
	}
	retentionManager := storage.NewRetentionManager(db, cfg.RetentionDays)

	// 3. Initialize Analytics Services
	metricsService := analytics.NewMetricsService(eventStore)
	funnelEngine := analytics.NewFunnelEngine(db)
	realtimeTracker := analytics.NewRealtimeTracker()
	seeder := analytics.NewSeeder(db, eventStore, replayStore)

	// Auto-seed if requested
	if *autoSeed {
		log.Printf("[MicroHog] Checking database for seeding...")
		var count int
		_ = db.QueryRow("SELECT COUNT(*) FROM events").Scan(&count)
		if count == 0 {
			log.Printf("[MicroHog] Seeding demo traffic...")
			if err := seeder.SeedData("default"); err != nil {
				log.Printf("[MicroHog] Seeding error: %v", err)
			} else {
				log.Printf("[MicroHog] Seeding completed.")
			}
		}
	}

	// 4. Initialize Ingestion Layer
	anonymizer := ingest.NewAnonymizer(cfg.SaltSecret)
	bufferPool := ingest.NewBufferPool(eventStore, 10000, 200, 500*time.Millisecond)
	defer bufferPool.Stop()

	collector := ingest.NewCollector(anonymizer, bufferPool, replayStore, realtimeTracker)

	// 5. Initialize API and HTTP Router
	apiHandler := server.NewAPIHandler(
		db,
		eventStore,
		replayStore,
		metricsService,
		funnelEngine,
		realtimeTracker,
		seeder,
	)

	staticFS := server.GetEmbeddedFS()
	router := server.NewRouter(&server.RouterConfig{
		Collector:       collector,
		APIHandler:      apiHandler,
		RealtimeTracker: realtimeTracker,
		StaticFileFS:    staticFS,
	})

	// 6. Start Background Retention Worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go retentionManager.StartWorker(ctx)

	// 7. Start HTTP Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second, // Accommodate SSE connections
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[MicroHog] Web dashboard ready at http://localhost:%d", cfg.Port)
		log.Printf("[MicroHog] Ingestion endpoint: http://localhost:%d/e", cfg.Port)
		log.Printf("[MicroHog] Tracker script: http://localhost:%d/microhog.js", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[MicroHog] Server error: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("[MicroHog] Shutting down gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[MicroHog] Server forced shutdown: %v", err)
	}

	log.Printf("[MicroHog] Stopped cleanly.")
}

func runSeed(args []string) {
	fs := flag.NewFlagSet("seed", flag.ExitOnError)
	dbPath := fs.String("db", "./microhog.db", "Path to SQLite database file")
	siteID := fs.String("site", "default", "Site identifier to seed")
	_ = fs.Parse(args)

	db, err := storage.OpenDB(*dbPath)
	if err != nil {
		log.Fatalf("[Seed] Failed to open database: %v", err)
	}
	defer db.Close()

	eventStore := storage.NewEventStore(db)
	replayStore, err := storage.NewReplayStore(db, eventStore)
	if err != nil {
		log.Fatalf("[Seed] Failed to initialize replay store: %v", err)
	}

	seeder := analytics.NewSeeder(db, eventStore, replayStore)
	log.Printf("[Seed] Seeding realistic demo data into %s (site: %s)...", *dbPath, *siteID)
	if err := seeder.SeedData(*siteID); err != nil {
		log.Fatalf("[Seed] Failed: %v", err)
	}
	log.Printf("[Seed] Success! 240+ sessions and full session replay generated.")
}

func runPrune(args []string) {
	fs := flag.NewFlagSet("prune", flag.ExitOnError)
	dbPath := fs.String("db", "./microhog.db", "Path to SQLite database file")
	days := fs.Int("days", 90, "Retention period in days")
	_ = fs.Parse(args)

	db, err := storage.OpenDB(*dbPath)
	if err != nil {
		log.Fatalf("[Prune] Failed to open database: %v", err)
	}
	defer db.Close()

	rm := storage.NewRetentionManager(db, *days)
	log.Printf("[Prune] Pruning records older than %d days from %s...", *days, *dbPath)
	deleted, err := rm.Prune(*days)
	if err != nil {
		log.Fatalf("[Prune] Failed: %v", err)
	}
	log.Printf("[Prune] Pruning completed. Removed %d expired records.", deleted)
}
