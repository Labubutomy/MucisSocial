package main

import (
	"database/sql"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	groupspb "github.com/Labubutomy/MucisSocial/services/groups/api"
	"github.com/Labubutomy/MucisSocial/services/groups/internal"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:password@postgres-groups:5432/music_social_groups?sslmode=disable")
	grpcPort := getEnv("GRPC_PORT", "9095")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("db ping failed: %v", err)
	}

	cfg := internal.LoadConfig()
	queueClient, queueClose, err := internal.NewPlaybackQueueClient(cfg.QueueServiceAddr)
	if err != nil {
		log.Fatalf("failed to connect queue service: %v", err)
	}
	defer queueClose()
	repo := internal.NewRepository(db)
	svc := internal.NewService(repo, cfg, queueClient)
	grpcSrv := internal.NewGRPCServer(svc)
	grpcServer := grpc.NewServer()
	groupspb.RegisterGroupsServiceServer(grpcServer, grpcSrv)

	grpcListener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("grpc listener failed: %v", err)
	}

	go func() {
		log.Printf("group gRPC service listening on %s", grpcPort)
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("grpc server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutdown signal received, stopping gRPC server")
	grpcServer.GracefulStop()
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
