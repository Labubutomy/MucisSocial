package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	queuepb "github.com/Labubutomy/MucisSocial/services/playback_queue/api"
	"github.com/Labubutomy/MucisSocial/services/playback_queue/internal"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Printf("env %s not set, using default: %s", key, defaultVal)
	return defaultVal
}

type appConfig struct {
	GRPCPort    string
	DatabaseURL string
}

func loadAppConfig() appConfig {
	return appConfig{
		GRPCPort:    getEnv("GRPC_PORT", "50056"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/playback_queue?sslmode=disable"),
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := loadAppConfig()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	repo := internal.NewRepository(db)
	service := internal.NewService(repo)
	grpcServer := grpc.NewServer()
	queuepb.RegisterPlaybackQueueServiceServer(grpcServer, internal.NewGRPCServer(service))
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	go func() {
		log.Printf("playback queue gRPC server listening on %s", cfg.GRPCPort)
		if serveErr := grpcServer.Serve(listener); serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
			log.Fatalf("gRPC server error: %v", serveErr)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received")

	log.Println("stopping gRPC server")
	grpcServer.GracefulStop()
	log.Println("gRPC server stopped")
}
