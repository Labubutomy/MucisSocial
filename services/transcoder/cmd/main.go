package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MusicSocial/transcoder/internal/broker"
	"github.com/MusicSocial/transcoder/internal/config"
	"github.com/MusicSocial/transcoder/internal/storage"
	"github.com/MusicSocial/transcoder/internal/tracks"
	"github.com/MusicSocial/transcoder/internal/transcoder"
)

func main() {
	cfg := config.Load()

	logger := log.New(os.Stdout, "[transcoder] ", log.LstdFlags|log.Lmicroseconds)

	minioClient, err := storage.NewMinIO(cfg.MinIO)
	if err != nil {
		logger.Fatalf("failed to init minio client: %v", err)
	}

	trackClient, err := tracks.NewTrackClient(&cfg.TrackService)
	if err != nil {
		logger.Fatalf("failed to init track service client: %v", err)
	}
	defer func() {
		if err := trackClient.Close(); err != nil {
			logger.Printf("failed to close track service client: %v", err)
		}
	}()

	worker := transcoder.NewFFmpegTranscoder(minioClient, trackClient, cfg.WorkDir, logger)

	consumer, err := broker.NewConsumer(cfg.Kafka, worker, logger)
	if err != nil {
		logger.Fatalf("failed to create consumer: %v", err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			logger.Printf("failed to close consumer: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Printf("starting consumer (topic=%s, group=%s)", cfg.Kafka.Topic, cfg.Kafka.GroupID)

	if err := consumer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Fatalf("consumer stopped with error: %v", err)
	}

	logger.Println("consumer stopped")
}
