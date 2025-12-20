package store

import (
	"sync"
	"testing"
	"time"
)

func TestGeoTopStore_Incr(t *testing.T) {
	store := NewInMemoryGeoTopStore(1000, 60*time.Second)

	store.Incr("u4pruyd", "track1", 1)
	store.Incr("u4pruyd", "track2", 1)
	store.Incr("u4pruyd", "track1", 1)

	tops := store.GetTopForGeohashes([]string{"u4pruyd"}, 10)

	if len(tops) != 2 {
		t.Fatalf("Expected 2 tracks, got %d", len(tops))
	}

	if tops[0].TrackID != "track1" || tops[0].Count != 2 {
		t.Errorf("Expected track1 with count 2, got %s with count %d", tops[0].TrackID, tops[0].Count)
	}

	if tops[1].TrackID != "track2" || tops[1].Count != 1 {
		t.Errorf("Expected track2 with count 1, got %s with count %d", tops[1].TrackID, tops[1].Count)
	}
}

func TestGeoTopStore_ConcurrentIncr(t *testing.T) {
	store := NewInMemoryGeoTopStore(1000, 60*time.Second)
	geohash := "u4pruyd"
	trackID := "track1"

	var wg sync.WaitGroup
	numGoroutines := 100
	incrementsPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				store.Incr(geohash, trackID, 1)
			}
		}()
	}

	wg.Wait()

	tops := store.GetTopForGeohashes([]string{geohash}, 10)
	expectedCount := int64(numGoroutines * incrementsPerGoroutine)

	if len(tops) != 1 {
		t.Fatalf("Expected 1 track, got %d", len(tops))
	}

	if tops[0].Count != expectedCount {
		t.Errorf("Expected count %d, got %d", expectedCount, tops[0].Count)
	}
}

func TestGeoTopStore_Snapshot_Restore(t *testing.T) {
	store1 := NewInMemoryGeoTopStore(1000, 60*time.Second)
	store1.Incr("u4pru", "track1", 10)
	store1.Incr("u4pru", "track2", 5)
	store1.Incr("u4prv", "track3", 7)

	snapshot := store1.Snapshot()

	store2 := NewInMemoryGeoTopStore(1000, 60*time.Second)
	store2.Restore(snapshot)

	tops1 := store1.GetTopForGeohashes([]string{"u4pru", "u4prv"}, 10)
	tops2 := store2.GetTopForGeohashes([]string{"u4pru", "u4prv"}, 10)

	if len(tops1) != len(tops2) {
		t.Fatalf("Restored store has different number of tracks: %d vs %d", len(tops2), len(tops1))
	}

	for i := range tops1 {
		if tops1[i].TrackID != tops2[i].TrackID || tops1[i].Count != tops2[i].Count {
			t.Errorf("Mismatch at index %d: expected %v, got %v", i, tops1[i], tops2[i])
		}
	}
}
