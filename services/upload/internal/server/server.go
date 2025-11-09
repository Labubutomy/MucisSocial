package server

import (
	"bytes"
	"fmt"
	"io"
	"log"

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
		artistIDs []string
		trackName string
		genre     string
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

	artistIDs = metadata.ArtistIds
	trackName = metadata.TrackName
	genre = metadata.Genre

	if len(artistIDs) == 0 {
		return fmt.Errorf("metadata must include at least one artist_id")
	}

	primaryArtist := artistIDs[0]

	log.Printf("Starting track upload: artist_ids=%v, track_name=%s, genre=%s", artistIDs, trackName, genre)

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

	ctx := stream.Context()
	trackID, err := s.trackClient.CreateTrack(ctx, trackName, artistIDs, genre, 0)
	if err != nil {
		return fmt.Errorf("failed to create track in track service: %w", err)
	}

	log.Printf("Track created in Track Service: track_id=%s", trackID)

	// Detect file extension from content
	extension, err := audio.DetectExtension(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to detect audio format: %w", err)
	}
	log.Printf("Detected audio format: %s", extension)

	reader := bytes.NewReader(buffer.Bytes())
	objectName, err := s.storage.UploadTrack(ctx, reader, primaryArtist, trackID, extension)
	if err != nil {
		return fmt.Errorf("failed to upload to storage: %w", err)
	}

	trackURL := fmt.Sprintf("http://minio:9000/%s/%s", s.config.MinIO.BucketName, objectName)

	log.Printf("Track uploaded to MinIO: %s", trackURL)

	transcoderTask := messaging.TranscoderTask{
		TrackID:  trackID,
		ArtistID: primaryArtist,
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
