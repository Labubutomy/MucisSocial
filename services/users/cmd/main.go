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

	"github.com/MucisSocial/user-service/internal/config"
	"github.com/MucisSocial/user-service/internal/handler"
	"github.com/MucisSocial/user-service/internal/repository"
	"github.com/MucisSocial/user-service/internal/service"
	pb "github.com/MucisSocial/user-service/proto/users/v1"
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

	logger.Info("Starting user service",
		zap.String("host", cfg.Server.Host),
		zap.String("port", cfg.Server.Port))

	// Setup database
	db, err := setupDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to setup database", zap.Error(err))
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(cfg.DatabaseDSN()); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Setup repositories
	userRepo := repository.NewUserRepository(db)
	searchHistoryRepo := repository.NewSearchHistoryRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// Setup services
	jwtService := service.NewJWTService(&cfg.JWT)
	authService := service.NewAuthService(userRepo, refreshTokenRepo, jwtService)
	userService := service.NewUserService(userRepo)
	searchHistoryService := service.NewSearchHistoryService(searchHistoryRepo)

	// Setup gRPC server
	grpcServer := grpc.NewServer()
	userServiceHandler := handler.NewUserServiceHandler(authService, userService, searchHistoryService)
	pb.RegisterUserServiceServer(grpcServer, userServiceHandler)

	// Enable reflection for development
	reflection.Register(grpcServer)

	// Start server
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		logger.Info("gRPC server starting", zap.String("addr", listener.Addr().String()))
		if err := grpcServer.Serve(listener); err != nil {
			logger.Error("gRPC server failed", zap.Error(err))
			cancel()
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("Shutting down server...")
	case <-ctx.Done():
		logger.Info("Server context cancelled")
	}

	// Graceful shutdown with timeout
	logger.Info("Stopping gRPC server...")
	grpcServer.GracefulStop()
	logger.Info("Server stopped")
}

func setupLogger(cfg config.LoggerConfig) *zap.Logger {
	var logger *zap.Logger
	var err error

	if cfg.Env == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		log.Fatal("Failed to setup logger:", err)
	}

	return logger
}

func setupDatabase(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxConns)
	db.SetMaxIdleConns(cfg.Database.MaxConns / 2)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func runMigrations(databaseURL string) error {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Get the path to migrations directory
	migrationsPath := "migrations"
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		// Try relative to executable
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %w", err)
		}
		migrationsPath = filepath.Join(filepath.Dir(execPath), "migrations")
		if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
			migrationsPath = "./migrations"
		}
	}

	migrationsURL := fmt.Sprintf("file://%s", migrationsPath)
	m, err := migrate.NewWithDatabaseInstance(
		migrationsURL,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
