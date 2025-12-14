package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOBackupStore struct {
	client     *minio.Client
	bucketName string

	// In-memory stores that we backup/restore
	trackStore       TrackStore
	userProfileStore UserProfileStore
	globalStatsStore GlobalStatsStore
}

type BackupData struct {
	Tracks       map[string]*models.Track       `json:"tracks"`
	UserProfiles map[string]*models.UserProfile `json:"user_profiles"`
	GlobalStats  map[string]int                 `json:"global_stats"`
	Timestamp    time.Time                      `json:"timestamp"`
}

func NewMinIOBackupStore(
	endpoint, accessKey, secretKey, bucketName string,
	trackStore TrackStore,
	userProfileStore UserProfileStore,
	globalStatsStore GlobalStatsStore,
) (*MinIOBackupStore, error) {

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("Created bucket: %s", bucketName)
	}

	backup := &MinIOBackupStore{
		client:           client,
		bucketName:       bucketName,
		trackStore:       trackStore,
		userProfileStore: userProfileStore,
		globalStatsStore: globalStatsStore,
	}

	return backup, nil
}

// LoadFromBackup loads data from MinIO into memory stores
func (s *MinIOBackupStore) LoadFromBackup() error {
	ctx := context.Background()
	objectName := "recommendations_backup.json"

	log.Printf("Loading backup from MinIO: %s/%s", s.bucketName, objectName)

	object, err := s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Printf("No backup found, starting with empty state: %v", err)
		return nil // Not an error - first time startup
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	var backup BackupData
	if err := json.Unmarshal(data, &backup); err != nil {
		return fmt.Errorf("failed to unmarshal backup: %w", err)
	}

	// Restore tracks
	if memTrackStore, ok := s.trackStore.(*InMemoryTrackStore); ok {
		for _, track := range backup.Tracks {
			memTrackStore.Upsert(track)
		}
		log.Printf("Restored %d tracks from backup", len(backup.Tracks))
	}

	// Restore user profiles
	if memUserStore, ok := s.userProfileStore.(*InMemoryUserProfileStore); ok {
		for _, profile := range backup.UserProfiles {
			memUserStore.Update(profile)
		}
		log.Printf("Restored %d user profiles from backup", len(backup.UserProfiles))
	}

	// Restore global stats
	if memStatsStore, ok := s.globalStatsStore.(*InMemoryGlobalStatsStore); ok {
		for trackID, count := range backup.GlobalStats {
			for i := 0; i < count; i++ {
				memStatsStore.IncrementPlayCount(trackID)
			}
		}
		log.Printf("Restored %d track play counts from backup", len(backup.GlobalStats))
	}

	log.Printf("Successfully loaded backup from %v", backup.Timestamp)
	return nil
}

// SaveBackup saves current in-memory state to MinIO
func (s *MinIOBackupStore) SaveBackup() error {
	ctx := context.Background()
	objectName := "recommendations_backup.json"

	backup := BackupData{
		Tracks:       make(map[string]*models.Track),
		UserProfiles: make(map[string]*models.UserProfile),
		GlobalStats:  make(map[string]int),
		Timestamp:    time.Now(),
	}

	// Collect tracks
	if memTrackStore, ok := s.trackStore.(*InMemoryTrackStore); ok {
		tracks := memTrackStore.GetAll()
		for _, track := range tracks {
			backup.Tracks[track.TrackID] = track
		}
	}

	// Collect user profiles
	if memUserStore, ok := s.userProfileStore.(*InMemoryUserProfileStore); ok {
		backup.UserProfiles = memUserStore.GetAll()
	}

	// Collect global stats
	if memStatsStore, ok := s.globalStatsStore.(*InMemoryGlobalStatsStore); ok {
		backup.GlobalStats = memStatsStore.GetAllPlayCounts()
	}

	// Serialize backup
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backup: %w", err)
	}

	// Upload to MinIO
	reader := bytes.NewReader(data)
	_, err = s.client.PutObject(ctx, s.bucketName, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	if err != nil {
		return fmt.Errorf("failed to upload backup: %w", err)
	}

	log.Printf("Successfully saved backup: %d tracks, %d users, %d stats",
		len(backup.Tracks), len(backup.UserProfiles), len(backup.GlobalStats))

	return nil
}

// StartPeriodicBackup starts a goroutine that saves backup every interval
func (s *MinIOBackupStore) StartPeriodicBackup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.SaveBackup(); err != nil {
					log.Printf("Periodic backup failed: %v", err)
				}
			}
		}
	}()
	log.Printf("Started periodic backup every %v", interval)
}
