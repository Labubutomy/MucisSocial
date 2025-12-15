package engine_test

import (
	"testing"
	"time"

	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/engine"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/engine/generators"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/engine/scorers"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/store"
)

func TestRecommendationEngineBasic(t *testing.T) {
	trackStore := store.NewInMemoryTrackStore()
	userProfileStore := store.NewInMemoryUserProfileStore()
	globalStatsStore := store.NewInMemoryGlobalStatsStore()

	now := time.Now().Unix()
	tracks := []*models.Track{
		{TrackID: "track1", ArtistID: "artist1", Genres: []string{"rock"}, ReleaseTS: now, IsExplicit: false},
		{TrackID: "track2", ArtistID: "artist1", Genres: []string{"rock", "indie"}, ReleaseTS: now, IsExplicit: false},
		{TrackID: "track3", ArtistID: "artist2", Genres: []string{"pop"}, ReleaseTS: now, IsExplicit: true},
		{TrackID: "track4", ArtistID: "artist2", Genres: []string{"pop"}, ReleaseTS: now, IsExplicit: false},
		{TrackID: "track5", ArtistID: "artist3", Genres: []string{"electronic"}, ReleaseTS: now, IsExplicit: false},
	}

	for _, track := range tracks {
		trackStore.Upsert(track)
		globalStatsStore.IncrementPlayCount(track.TrackID)
	}

	profile := models.NewUserProfile("user1")
	profile.GenreListenCount["rock"] = 10
	profile.GenreListenCount["indie"] = 5
	profile.ArtistListenCount["artist1"] = 8
	profile.ListenedTracks["track1"] = struct{}{}
	userProfileStore.Update(profile)

	genreGen := generators.NewTopGenresGenerator(trackStore, globalStatsStore, 5, 50)
	artistGen := generators.NewTopArtistsGenerator(trackStore, globalStatsStore, 5, 50)
	fallbackGen := generators.NewFallbackGenerator(trackStore, globalStatsStore, 100)

	scorer := scorers.NewHeuristicScorer(trackStore, globalStatsStore)

	recEngine := engine.NewRecommendationEngine(
		[]engine.CandidateGenerator{genreGen, artistGen, fallbackGen},
		scorer,
		trackStore,
	)

	req := &engine.RecommendRequest{
		UserID:  "user1",
		Limit:   10,
		Filters: nil,
	}

	result := recEngine.Recommend(profile, req)

	for _, trackID := range result {
		if trackID == "track1" {
			t.Error("Result should not include already listened track1")
		}
	}

	found := false
	for _, trackID := range result {
		if trackID == "track2" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Result should include track2 (matching genre and artist)")
	}
}

func TestRecommendationEngineExplicitFilter(t *testing.T) {
	trackStore := store.NewInMemoryTrackStore()
	globalStatsStore := store.NewInMemoryGlobalStatsStore()

	now := time.Now().Unix()
	trackStore.Upsert(&models.Track{TrackID: "explicit", ArtistID: "a1", Genres: []string{"rock"}, ReleaseTS: now, IsExplicit: true})
	trackStore.Upsert(&models.Track{TrackID: "clean", ArtistID: "a1", Genres: []string{"rock"}, ReleaseTS: now, IsExplicit: false})
	globalStatsStore.IncrementPlayCount("explicit")
	globalStatsStore.IncrementPlayCount("clean")

	profile := models.NewUserProfile("user1")
	profile.GenreListenCount["rock"] = 5

	genreGen := generators.NewTopGenresGenerator(trackStore, globalStatsStore, 5, 50)
	fallbackGen := generators.NewFallbackGenerator(trackStore, globalStatsStore, 100)
	scorer := scorers.NewHeuristicScorer(trackStore, globalStatsStore)

	recEngine := engine.NewRecommendationEngine(
		[]engine.CandidateGenerator{genreGen, fallbackGen},
		scorer,
		trackStore,
	)

	req := &engine.RecommendRequest{
		UserID: "user1",
		Limit:  10,
		Filters: &engine.FilterOptions{
			ExcludeExplicit: true,
		},
	}

	result := recEngine.Recommend(profile, req)

	for _, trackID := range result {
		if trackID == "explicit" {
			t.Error("Result should not include explicit track when filter is enabled")
		}
	}
}

func TestRecommendationEngineGenreFilter(t *testing.T) {
	trackStore := store.NewInMemoryTrackStore()
	globalStatsStore := store.NewInMemoryGlobalStatsStore()

	now := time.Now().Unix()
	trackStore.Upsert(&models.Track{TrackID: "rock1", ArtistID: "a1", Genres: []string{"rock"}, ReleaseTS: now, IsExplicit: false})
	trackStore.Upsert(&models.Track{TrackID: "pop1", ArtistID: "a1", Genres: []string{"pop"}, ReleaseTS: now, IsExplicit: false})
	globalStatsStore.IncrementPlayCount("rock1")
	globalStatsStore.IncrementPlayCount("pop1")

	profile := models.NewUserProfile("user1")

	fallbackGen := generators.NewFallbackGenerator(trackStore, globalStatsStore, 100)
	scorer := scorers.NewHeuristicScorer(trackStore, globalStatsStore)

	recEngine := engine.NewRecommendationEngine(
		[]engine.CandidateGenerator{fallbackGen},
		scorer,
		trackStore,
	)

	req := &engine.RecommendRequest{
		UserID: "user1",
		Limit:  10,
		Filters: &engine.FilterOptions{
			Genres: []string{"rock"},
		},
	}

	result := recEngine.Recommend(profile, req)

	for _, trackID := range result {
		if trackID == "pop1" {
			t.Error("Result should not include pop track when filtering for rock")
		}
	}

	found := false
	for _, trackID := range result {
		if trackID == "rock1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Result should include rock1 track")
	}
}

func TestScorerInterface(t *testing.T) {
	trackStore := store.NewInMemoryTrackStore()
	globalStatsStore := store.NewInMemoryGlobalStatsStore()

	now := time.Now().Unix()
	trackStore.Upsert(&models.Track{TrackID: "t1", ArtistID: "a1", Genres: []string{"rock"}, ReleaseTS: now, IsExplicit: false})
	globalStatsStore.IncrementPlayCount("t1")

	var scorer engine.Scorer

	scorer = scorers.NewHeuristicScorer(trackStore, globalStatsStore)
	if scorer.Name() != "heuristic_v1" {
		t.Errorf("Expected heuristic_v1, got %s", scorer.Name())
	}

	scorer = scorers.NewMLScorer("test/path")
	if scorer.Name() != "ml_ranker_v1" {
		t.Errorf("Expected ml_ranker_v1, got %s", scorer.Name())
	}
}

func TestCandidateGeneratorInterface(t *testing.T) {
	trackStore := store.NewInMemoryTrackStore()
	globalStatsStore := store.NewInMemoryGlobalStatsStore()

	var gen engine.CandidateGenerator

	gen = generators.NewTopGenresGenerator(trackStore, globalStatsStore, 5, 50)
	if gen.Name() != "top_genres" {
		t.Errorf("Expected top_genres, got %s", gen.Name())
	}

	gen = generators.NewTopArtistsGenerator(trackStore, globalStatsStore, 5, 50)
	if gen.Name() != "top_artists" {
		t.Errorf("Expected top_artists, got %s", gen.Name())
	}

	gen = generators.NewFallbackGenerator(trackStore, globalStatsStore, 100)
	if gen.Name() != "fallback_global" {
		t.Errorf("Expected fallback_global, got %s", gen.Name())
	}
}
