package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/MucisSocial/upload/internal/config"
	"github.com/MucisSocial/upload/internal/messaging"
	"github.com/MucisSocial/upload/internal/server"
	"github.com/MucisSocial/upload/internal/storage"
	"github.com/MucisSocial/upload/internal/tracks"
	pb "github.com/MucisSocial/upload/proto"
	"google.golang.org/grpc"
)

func main() {
	log.Println("Starting Upload Service...")

	cfg := config.Load()

	// MinIO
	minioStorage, err := storage.NewMinIOStorage(&cfg.MinIO)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO storage: %v", err)
	}
	log.Println("MinIO storage initialized successfully")

	// Redpanda
	producer, err := messaging.NewProducer(&cfg.Redpanda)
	if err != nil {
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}
	defer producer.Close()
	log.Println("Kafka producer initialized successfully")

	// Track Service
	trackClient, err := tracks.NewTrackClient(&cfg.Tracks)
	if err != nil {
		log.Fatalf("Failed to initialize Track Service client: %v", err)
	}
	defer trackClient.Close()
	log.Println("Track Service client initialized successfully")

	// gRPC input server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(100*1024*1024),
		grpc.MaxSendMsgSize(100*1024*1024),
	)
	uploadServer := server.NewUploadServer(cfg, minioStorage, producer, trackClient)
	pb.RegisterUploadServiceServer(grpcServer, uploadServer)

	addr := fmt.Sprintf(":%s", cfg.Server.GRPCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	log.Printf("Upload service is listening on %s", addr)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Upload Service...")
	grpcServer.GracefulStop()
	log.Println("Upload Service stopped")
}
