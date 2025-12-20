package models

import "testing"

func TestAggregateProfiles(t *testing.T) {
	user1 := &UserProfile{
		UserID:            "user1",
		GenreListenCount:  map[string]int{"rock": 10, "pop": 5},
		ArtistListenCount: map[string]int{"artist1": 15},
		ListenedTracks:    map[string]struct{}{"track1": {}, "track2": {}},
	}

	user2 := &UserProfile{
		UserID:            "user2",
		GenreListenCount:  map[string]int{"rock": 8, "jazz": 12},
		ArtistListenCount: map[string]int{"artist1": 5},
		ListenedTracks:    map[string]struct{}{"track2": {}, "track3": {}},
	}

	profiles := []*UserProfile{user1, user2}
	groupProfile := AggregateProfiles(profiles)

	if groupProfile.TotalUsers != 2 {
		t.Errorf("Expected TotalUsers = 2, got %d", groupProfile.TotalUsers)
	}

	if groupProfile.GenreListenCount["rock"] != 18 {
		t.Errorf("Expected rock count = 18, got %d", groupProfile.GenreListenCount["rock"])
	}

	if groupProfile.ListenedTracks["track2"] != 2 {
		t.Errorf("Expected track2 heard by 2 users, got %d", groupProfile.ListenedTracks["track2"])
	}
}

func TestGroupProfile_HasListenedByMajority(t *testing.T) {
	group := &GroupProfile{
		TotalUsers:     4,
		ListenedTracks: map[string]int{"track1": 1, "track2": 2, "track3": 3},
	}

	if group.HasListenedByMajority("track1") {
		t.Error("track1 should not be majority (1/4)")
	}

	if group.HasListenedByMajority("track2") {
		t.Error("track2 should not be majority (2/4 = 50%)")
	}

	if !group.HasListenedByMajority("track3") {
		t.Error("track3 should be majority (3/4 > 50%)")
	}
}
