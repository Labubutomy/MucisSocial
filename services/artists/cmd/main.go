package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/MucisSocial/artist-service/internal/config"
	"github.com/MucisSocial/artist-service/internal/handler"
	pb "github.com/MucisSocial/artist-service/internal/pb/artists/v1"
	"github.com/MucisSocial/artist-service/internal/repository"
	"github.com/MucisSocial/artist-service/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Setup logger
	logger := setupLogger(cfg.Logger)
	defer logger.Sync()

	logger.Info("Starting artist service",
		zap.String("host", cfg.Server.Host),
		zap.String("port", cfg.Server.Port))

	// Setup database
	db, err := setupDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to setup database", zap.Error(err))
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(cfg); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Initialize repository layer
	artistRepo := repository.NewArtistRepository(db)

	// Initialize service layer
	artistService := service.NewArtistService(artistRepo)

	// Initialize handler layer
	artistHandler := handler.NewArtistServiceHandler(artistService)

	// Setup gRPC server
	server := grpc.NewServer()
	pb.RegisterArtistServiceServer(server, artistHandler)

	// Enable reflection for development
	reflection.Register(server)

	// Start server
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("Artist service is listening", zap.String("address", listener.Addr().String()))
		if err := server.Serve(listener); err != nil {
			logger.Error("Server failed to serve", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down artist service...")

	// Graceful shutdown with timeout
	done := make(chan bool, 1)
	go func() {
		server.GracefulStop()
		done <- true
	}()

	select {
	case <-done:
		logger.Info("Artist service stopped gracefully")
	case <-time.After(30 * time.Second):
		logger.Warn("Force stopping artist service")
		server.Stop()
	}
}

func setupLogger(cfg config.LoggerConfig) *zap.Logger {
	var logger *zap.Logger
	var err error

	if cfg.Format == "json" {
		config := zap.NewProductionConfig()
		if err := config.Level.UnmarshalText([]byte(cfg.Level)); err == nil {
			logger, err = config.Build()
		} else {
			logger, err = zap.NewProduction()
		}
	} else {
		config := zap.NewDevelopmentConfig()
		if err := config.Level.UnmarshalText([]byte(cfg.Level)); err == nil {
			logger, err = config.Build()
		} else {
			logger, err = zap.NewDevelopment()
		}
	}

	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	return logger
}

func setupDatabase(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.Database.MaxConns)
	db.SetMaxIdleConns(cfg.Database.MaxConns / 2)
	db.SetConnMaxLifetime(time.Hour)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func runMigrations(cfg *config.Config) error {
	db, err := sql.Open("postgres", cfg.DatabaseConnectionString())
	if err != nil {
		return fmt.Errorf("failed to connect to database for migrations: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Get the directory where the binary is located
	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exPath := filepath.Dir(ex)

	// Try different migration paths
	migrationPaths := []string{
		filepath.Join(exPath, "migrations"),
		"migrations",
		"./migrations",
	}

	var migrationPath string
	for _, path := range migrationPaths {
		if _, err := os.Stat(path); err == nil {
			migrationPath = path
			break
		}
	}

	if migrationPath == "" {
		return fmt.Errorf("migrations directory not found in any of: %v", migrationPaths)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
