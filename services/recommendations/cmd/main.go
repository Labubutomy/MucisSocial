package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/api"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/bootstrap"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/config"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/consumer"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/engine"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/engine/generators"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/engine/scorers"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/store"
)

func main() {
	cfg := config.Load()

	// Initialize in-memory stores with MinIO backup
	log.Println("Using in-memory storage with MinIO backup")
	trackStore := store.NewInMemoryTrackStore()
	userProfileStore := store.NewInMemoryUserProfileStore()
	globalStatsStore := store.NewInMemoryGlobalStatsStore()

	// Initialize MinIO backup
	backupStore, err := store.NewMinIOBackupStore(
		cfg.MinIOEndpoint,
		cfg.MinIOAccessKey,
		cfg.MinIOSecretKey,
		cfg.MinIOBucketName,
		trackStore,
		userProfileStore,
		globalStatsStore,
	)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO backup: %v", err)
	}

	// Load existing data from backup
	if err := backupStore.LoadFromBackup(); err != nil {
		log.Printf("Failed to load backup (will start fresh): %v", err)
	}

	// Parse backup interval
	backupInterval, err := time.ParseDuration(cfg.BackupInterval)
	if err != nil {
		log.Printf("Invalid backup interval '%s', using 5m: %v", cfg.BackupInterval, err)
		backupInterval = 5 * time.Minute
	}

	// Start periodic backup
	backupStore.StartPeriodicBackup(backupInterval)

	// Setup graceful backup on shutdown
	defer func() {
		log.Println("Saving final backup before shutdown...")
		if err := backupStore.SaveBackup(); err != nil {
			log.Printf("Failed to save final backup: %v", err)
		}
	}()

	// Initialize candidate generators
	genreGenerator := generators.NewTopGenresGenerator(trackStore, globalStatsStore, 5, 50)
	artistGenerator := generators.NewTopArtistsGenerator(trackStore, globalStatsStore, 5, 50)
	fallbackGenerator := generators.NewFallbackGenerator(trackStore, globalStatsStore, 100)

	// Initialize scorer
	// NOTE: This is the ML replacement point. To use ML, replace with:
	// scorer := scorers.NewMLScorer("path/to/model")
	scorer := scorers.NewHeuristicScorer(trackStore, globalStatsStore)

	// Build recommendation engine
	recEngine := engine.NewRecommendationEngine(
		[]engine.CandidateGenerator{genreGenerator, artistGenerator, fallbackGenerator},
		scorer,
		trackStore,
	)

	// Initialize Kafka consumer
	eventConsumer := consumer.NewEventConsumer(
		cfg.KafkaBrokers,
		cfg.TrackEventsTopic,
		cfg.ListeningEventsTopic,
		trackStore,
		userProfileStore,
		globalStatsStore,
		cfg.TracksServiceURL,
	)

	// Start consuming events in background
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := eventConsumer.Start(ctx); err != nil {
			log.Printf("Event consumer error: %v", err)
		}
	}()

	// Bootstrap: load existing tracks from tracks service
	if cfg.BootstrapEnabled && cfg.TracksServiceURL != "" {
		bootstrapper := bootstrap.NewBootstrapper(cfg.TracksServiceURL, trackStore)
		go func() {
			if err := bootstrapper.Run(ctx); err != nil {
				log.Printf("Bootstrap error: %v", err)
			}
		}()
	}

	// Initialize HTTP server
	server := api.NewServer(cfg.HTTPPort, recEngine, userProfileStore, trackStore, globalStatsStore)

	// Add ingest endpoints for testing without Kafka
	server.AddIngestEndpoints(trackStore, userProfileStore, globalStatsStore)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Printf("Recommendation service started on port %s", cfg.HTTPPort)

	<-quit
	log.Println("Shutting down...")
	cancel()
	eventConsumer.Close()
}
