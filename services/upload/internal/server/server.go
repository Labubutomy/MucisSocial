package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"path/filepath"

	"github.com/MucisSocial/upload/internal/audio"
	"github.com/MucisSocial/upload/internal/config"
	"github.com/MucisSocial/upload/internal/messaging"
	"github.com/MucisSocial/upload/internal/storage"
	"github.com/MucisSocial/upload/internal/tracks"
	pb "github.com/MucisSocial/upload/proto"
)

type UploadServer struct {
	pb.UnimplementedUploadServiceServer
	storage     *storage.MinIOStorage
	producer    *messaging.Producer
	trackClient tracks.Client
	config      *config.Config
}

func NewUploadServer(cfg *config.Config, storage *storage.MinIOStorage, producer *messaging.Producer, trackClient tracks.Client) *UploadServer {
	return &UploadServer{
		storage:     storage,
		producer:    producer,
		trackClient: trackClient,
		config:      cfg,
	}
}

func (s *UploadServer) UploadTrack(stream pb.UploadService_UploadTrackServer) error {
	var (
		artistID  string
		trackName string
		buffer    bytes.Buffer
	)

	firstMsg, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive metadata: %w", err)
	}

	metadata := firstMsg.GetMetadata()
	if metadata == nil {
		return fmt.Errorf("first message must contain metadata")
	}

	artistID = metadata.ArtistId
	trackName = metadata.TrackName

	log.Printf("Starting track upload: artist_id=%s, track_name=%s", artistID, trackName)

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive chunk: %w", err)
		}

		chunk := msg.GetChunk()
		if chunk == nil {
			continue
		}

		if _, err := buffer.Write(chunk); err != nil {
			return fmt.Errorf("failed to write chunk to buffer: %w", err)
		}
	}

	log.Printf("Received track data: %d bytes", buffer.Len())

	duration, err := audio.GetDuration(buffer.Bytes())
	if err != nil {
		log.Printf("Warning: failed to get track duration: %v", err)
		duration = 0 // unknown duration
	}

	log.Printf("Track duration: %d seconds", duration)

	ctx := stream.Context()
	trackID, err := s.trackClient.CreateTrack(ctx, artistID, trackName, duration)
	if err != nil {
		return fmt.Errorf("failed to create track in track service: %w", err)
	}

	log.Printf("Track created in Track Service: track_id=%s", trackID)

	reader := bytes.NewReader(buffer.Bytes())
	extension := filepath.Ext(trackName)
	objectName, err := s.storage.UploadTrack(ctx, reader, artistID, trackID, extension)
	if err != nil {
		return fmt.Errorf("failed to upload to storage: %w", err)
	}

	trackURL := fmt.Sprintf("http://minio:9000/%s/%s", s.config.MinIO.BucketName, objectName)

	log.Printf("Track uploaded to MinIO: %s", trackURL)

	if err := s.trackClient.UpdateTrackLocation(ctx, trackID, trackURL); err != nil {
		return fmt.Errorf("failed to update track location in track service: %w", err)
	}
	log.Printf("Track location updated in Track Service: %s", trackURL)

	transcoderTask := messaging.TranscoderTask{
		TrackID:  trackID,
		ArtistID: artistID,
		TrackURL: trackURL,
	}

	if err := s.producer.SendTranscoderTask(ctx, transcoderTask); err != nil {
		log.Printf("Failed to send transcoder task: %v", err)
	} else {
		log.Printf("Transcoder task sent for track: %s", trackID)
	}

	response := &pb.UploadTrackResponse{
		Success: true,
		Message: "Track uploaded successfully",
		TrackId: trackID,
	}

	if err := stream.SendAndClose(response); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	log.Printf("Track upload completed successfully: %s", trackID)
	return nil
}
