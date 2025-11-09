package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/MusicSocial/transcoder/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIO struct {
	client     *minio.Client
	bucketName string
	endpoint   string
}

func NewMinIO(cfg config.MinIOConfig) (*MinIO, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &MinIO{
		client:     client,
		bucketName: cfg.BucketName,
		endpoint:   cfg.Endpoint,
	}, nil
}

func (m *MinIO) Endpoint() string {
	return m.endpoint
}
func (m *MinIO) Bucket() string {
	return m.bucketName
}

func (m *MinIO) DownloadToFile(ctx context.Context, bucket, objectKey, destPath string) error {
	reader, err := m.client.GetObject(ctx, bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object %s: %w", objectKey, err)
	}
	defer reader.Close()

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directories for %s: %w", destPath, err)
	}

	dest, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", destPath, err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, reader); err != nil {
		return fmt.Errorf("failed to copy object data to %s: %w", destPath, err)
	}

	return nil
}

func (m *MinIO) UploadFile(ctx context.Context, bucket, objectKey, filePath, contentType string) error {
	if contentType == "" {
		contentType = contentTypeFor(filePath)
	}

	_, err := m.client.FPutObject(ctx, bucket, objectKey, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload %s to %s: %w", filePath, objectKey, err)
	}

	return nil
}

func (m *MinIO) UploadBytes(ctx context.Context, bucket, objectKey string, data []byte, contentType string) error {
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := m.client.PutObject(ctx, bucket, objectKey, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object %s: %w", objectKey, err)
	}
	return nil
}

func (m *MinIO) UploadJSON(ctx context.Context, bucket, objectKey string, payload interface{}) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal json for %s: %w", objectKey, err)
	}
	return m.UploadBytes(ctx, bucket, objectKey, data, "application/json")
}

func (m *MinIO) UploadDirectory(ctx context.Context, bucket, prefix, dir string) error {
	return filepath.WalkDir(dir, func(entryPath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(dir, entryPath)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", entryPath, err)
		}

		objectKey := path.Join(prefix, filepath.ToSlash(rel))
		return m.UploadFile(ctx, bucket, objectKey, entryPath, contentTypeFor(entryPath))
	})
}

func contentTypeFor(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return "application/json"
	case ".m3u8":
		return "application/vnd.apple.mpegurl"
	case ".mp4":
		return "video/mp4"
	case ".m4s":
		return "video/mp4"
	default:
		if c := mime.TypeByExtension(ext); c != "" {
			return c
		}
		return "application/octet-stream"
	}
}
