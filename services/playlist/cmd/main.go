package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/Labubutomy/MucisSocial/services/playlist/api"
	"github.com/Labubutomy/MucisSocial/services/playlist/internal"
	"github.com/Labubutomy/MucisSocial/services/playlist/pkg"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load config
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/playlists_db?sslmode=disable")
	httpPort := getEnv("PORT", "8080")
	grpcPort := getEnv("GRPC_PORT", "50054")

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
	httpHandler := internal.NewHandler(service)
	grpcHandler := internal.NewGRPCHandler(service)

	// Setup Gin router
	router := gin.Default()
	router.Use(pkg.SimpleLoggingMiddleware())
	httpHandler.RegisterRoutes(router)

	// Setup HTTP server
	httpServer := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Setup gRPC server
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPlaylistServiceServer(grpcServer, grpcHandler)

	// Enable reflection for testing with tools like grpcurl
	reflection.Register(grpcServer)

	// Start HTTP server
	go func() {
		log.Printf("HTTP server starting on port %s", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start gRPC server
	go func() {
		log.Printf("gRPC server starting on port %s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	} else {
		log.Println("HTTP server stopped")
	}

	// Shutdown gRPC server
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Println("gRPC server stopped gracefully")
	case <-time.After(5 * time.Second):
		log.Println("Force stopping gRPC server...")
		grpcServer.Stop()
	}

	log.Println("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
