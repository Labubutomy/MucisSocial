package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Labubutomy/MucisSocial/services/tracks/internal"
	"github.com/Labubutomy/MucisSocial/services/tracks/pkg"
	_ "github.com/lib/pq"
)

func main() {
	// Load config
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/tracks_db?sslmode=disable")
	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	port := getEnv("PORT", "8080")

	// Connect to DB
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Check connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Connected to database")

	// Initialize layers
	repo := internal.NewRepository(db)
	service := internal.NewService(repo)
	handler := internal.NewHandler(service)

	// Start Kafka consumer
	consumer := internal.NewEventConsumer(service, kafkaBrokers)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Println("Starting Kafka consumer...")
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Kafka consumer error: %v", err)
		}
	}()

	// Setup HTTP server
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      pkg.LoggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Server error:", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server shutdown error:", err)
	}

	cancel() // Stop Kafka consumer
	log.Println("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
