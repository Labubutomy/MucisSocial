package storage

import (
	"context"
	"fmt"
	"io"
	"mime"
	"strings"

	"github.com/MusicSocial/upload/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	client     *minio.Client
	bucketName string
}

func NewMinIOStorage(cfg *config.MinIOConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &MinIOStorage{
		client:     client,
		bucketName: cfg.BucketName,
	}, nil
}

func (s *MinIOStorage) UploadTrack(ctx context.Context, reader io.Reader, artistID, trackID, extension string) (string, error) {
	if extension != "" {
		if !strings.HasPrefix(extension, ".") {
			extension = "." + extension
		}
	} else {
		return "", fmt.Errorf("extension is required")
	}

	objectName := fmt.Sprintf("%s/%s/original/original%s", artistID, trackID, extension)

	contentType := mime.TypeByExtension(extension)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.client.PutObject(
		ctx,
		s.bucketName,
		objectName,
		reader,
		-1, // unknown size
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload track: %w", err)
	}

	return objectName, nil
}
