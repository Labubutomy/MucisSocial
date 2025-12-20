package api

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/engine"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/models"
	"github.com/Labubutomy/MucisSocial/services/recommendations/internal/store"
	"github.com/gin-gonic/gin"
)

// Server handles HTTP requests
type Server struct {
	port             string
	engine           *engine.RecommendationEngine
	userProfileStore store.UserProfileStore
	trackStore       store.TrackStore
	globalStatsStore store.GlobalStatsStore
	searchQueryStore store.SearchQueryStore
	geoTopStore      store.GeoTopStore
	enableGeoTop     bool
	geohashPrecision int
	defaultRadiusM   int
	router           *gin.Engine
}

// NewServer creates a new HTTP server
func NewServer(
	port string,
	recEngine *engine.RecommendationEngine,
	userProfileStore store.UserProfileStore,
	trackStore store.TrackStore,
	globalStatsStore store.GlobalStatsStore,
	searchQueryStore store.SearchQueryStore,
	geoTopStore store.GeoTopStore,
	enableGeoTop bool,
	geohashPrecision int,
	defaultRadiusM int,
) *Server {
	s := &Server{
		port:             port,
		engine:           recEngine,
		userProfileStore: userProfileStore,
		trackStore:       trackStore,
		globalStatsStore: globalStatsStore,
		searchQueryStore: searchQueryStore,
		geoTopStore:      geoTopStore,
		enableGeoTop:     enableGeoTop,
		geohashPrecision: geohashPrecision,
		defaultRadiusM:   defaultRadiusM,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	gin.SetMode(gin.ReleaseMode)
	s.router = gin.New()
	s.router.Use(gin.Recovery())
	s.router.Use(gin.Logger())

	s.router.GET("/health", s.healthHandler)
	s.router.POST("/recommendations", s.recommendationsHandler)
	s.router.POST("/recommendations/group", s.groupRecommendationsHandler)
	s.router.GET("/charts/top", s.chartsTopHandler)
	s.router.GET("/charts/top/nearby", s.chartsTopNearbyHandler)
	s.router.GET("/tracks/new", s.newReleasesHandler)
	s.router.GET("/users/:user_id/taste", s.userTasteHandler)
	s.router.GET("/search/trending", s.trendingQueriesHandler)
	s.router.POST("/search/query", s.recordSearchQueryHandler)
	s.router.GET("/debug/tracks", s.debugTracksHandler)
	s.router.GET("/debug/users/:user_id", s.debugUserHandler)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.router.Run(":" + s.port)
}

// RecommendationRequest represents the API request
type RecommendationRequest struct {
	UserID  string                 `json:"user_id" binding:"required"`
	Limit   int                    `json:"limit"`
	Filters *RecommendationFilters `json:"filters"`
}

// RecommendationFilters represents optional filters
type RecommendationFilters struct {
	ExcludeExplicit bool     `json:"exclude_explicit"`
	Genres          []string `json:"genres"`
	FreshDays       int      `json:"fresh_days"`
}

// RecommendationResponse represents the API response
type RecommendationResponse struct {
	Tracks []string `json:"tracks"`
}

// GroupRecommendationRequest represents a group recommendation request
type GroupRecommendationRequest struct {
	UserIDs []string               `json:"user_ids" binding:"required,min=1"`
	Limit   int                    `json:"limit"`
	Filters *RecommendationFilters `json:"filters"`
}

// GroupRecommendationResponse represents the group recommendation API response
type GroupRecommendationResponse struct {
	Tracks      []string `json:"tracks"`
	GroupSize   int      `json:"group_size"`
	UserIDs     []string `json:"user_ids"`
	Description string   `json:"description"`
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) recommendationsHandler(c *gin.Context) {
	var req RecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	userProfile := s.userProfileStore.GetOrCreate(req.UserID)

	var filterOpts *engine.FilterOptions
	if req.Filters != nil {
		filterOpts = &engine.FilterOptions{
			ExcludeExplicit: req.Filters.ExcludeExplicit,
			Genres:          req.Filters.Genres,
			FreshDays:       req.Filters.FreshDays,
		}
	}

	recReq := &engine.RecommendRequest{
		UserID:  req.UserID,
		Limit:   req.Limit,
		Filters: filterOpts,
	}

	tracks := s.engine.Recommend(userProfile, recReq)

	c.JSON(http.StatusOK, RecommendationResponse{
		Tracks: tracks,
	})
}

func (s *Server) groupRecommendationsHandler(c *gin.Context) {
	var req GroupRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.UserIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_ids cannot be empty"})
		return
	}

	if len(req.UserIDs) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "maximum 50 users allowed in a group"})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Fetch profiles for all users in the group
	profiles := make([]*models.UserProfile, 0, len(req.UserIDs))
	for _, userID := range req.UserIDs {
		profile := s.userProfileStore.GetOrCreate(userID)
		profiles = append(profiles, profile)
	}

	var filterOpts *engine.FilterOptions
	if req.Filters != nil {
		filterOpts = &engine.FilterOptions{
			ExcludeExplicit: req.Filters.ExcludeExplicit,
			Genres:          req.Filters.Genres,
			FreshDays:       req.Filters.FreshDays,
		}
	}

	groupReq := &engine.GroupRecommendRequest{
		UserIDs: req.UserIDs,
		Limit:   req.Limit,
		Filters: filterOpts,
	}

	tracks := s.engine.RecommendForGroup(profiles, groupReq)

	c.JSON(http.StatusOK, GroupRecommendationResponse{
		Tracks:      tracks,
		GroupSize:   len(req.UserIDs),
		UserIDs:     req.UserIDs,
		Description: "Group recommendations that balance preferences across all users in the room",
	})
}

func (s *Server) debugTracksHandler(c *gin.Context) {
	tracks := s.trackStore.GetAll()
	c.JSON(http.StatusOK, gin.H{
		"count":  len(tracks),
		"tracks": tracks,
	})
}

func (s *Server) debugUserHandler(c *gin.Context) {
	userID := c.Param("user_id")
	profile, ok := s.userProfileStore.Get(userID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	listenedTracks := make([]string, 0, len(profile.ListenedTracks))
	for trackID := range profile.ListenedTracks {
		listenedTracks = append(listenedTracks, trackID)
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":             profile.UserID,
		"genre_listen_count":  profile.GenreListenCount,
		"artist_listen_count": profile.ArtistListenCount,
		"listened_tracks":     listenedTracks,
	})
}

// ChartTrack represents a track in the charts
type ChartTrack struct {
	TrackID   string `json:"track_id"`
	Position  int    `json:"position"`
	PlayCount int    `json:"play_count"`
}

// ChartsResponse represents the charts API response
type ChartsResponse struct {
	Period    string       `json:"period"`
	UpdatedAt string       `json:"updated_at"`
	Tracks    []ChartTrack `json:"tracks"`
}

func (s *Server) chartsTopHandler(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Get top tracks for current month (cached)
	trackIDs := s.globalStatsStore.GetTopTracksForMonth(limit)

	// Build response with positions and play counts
	tracks := make([]ChartTrack, 0, len(trackIDs))
	for i, trackID := range trackIDs {
		playCount := s.globalStatsStore.GetMonthlyPlayCount(trackID)
		tracks = append(tracks, ChartTrack{
			TrackID:   trackID,
			Position:  i + 1,
			PlayCount: playCount,
		})
	}

	c.JSON(http.StatusOK, ChartsResponse{
		Period:    "month",
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Tracks:    tracks,
	})
}

// NewReleasesResponse represents the new releases API response
type NewReleasesResponse struct {
	Tracks    []string `json:"tracks"`
	UpdatedAt string   `json:"updated_at"`
}

func (s *Server) newReleasesHandler(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	days := 30 // Default: last 30 days
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	// Get new releases (sorted by release date, most recent first)
	trackIDs := s.trackStore.GetNewReleases(limit, days)

	c.JSON(http.StatusOK, NewReleasesResponse{
		Tracks:    trackIDs,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	})
}

// UserTasteResponse represents the user taste statistics API response
type UserTasteResponse struct {
	UserID     string       `json:"user_id"`
	TopGenres  []GenreStat  `json:"top_genres"`
	TopArtists []ArtistStat `json:"top_artists"`
}

type GenreStat struct {
	Genre string `json:"genre"`
	Count int    `json:"count"`
}

type ArtistStat struct {
	ArtistID string `json:"artist_id"`
	Count    int    `json:"count"`
}

func (s *Server) userTasteHandler(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	profile, ok := s.userProfileStore.Get(userID)
	if !ok {
		// Return empty taste if user not found
		c.JSON(http.StatusOK, UserTasteResponse{
			UserID:     userID,
			TopGenres:  []GenreStat{},
			TopArtists: []ArtistStat{},
		})
		return
	}

	// Get top genres (sorted by count, descending)
	type genreCount struct {
		genre string
		count int
	}
	genres := make([]genreCount, 0, len(profile.GenreListenCount))
	for genre, count := range profile.GenreListenCount {
		genres = append(genres, genreCount{genre, count})
	}
	sort.Slice(genres, func(i, j int) bool {
		return genres[i].count > genres[j].count
	})

	topGenres := make([]GenreStat, 0, 10)
	maxGenres := 10
	if maxGenres > len(genres) {
		maxGenres = len(genres)
	}
	for i := 0; i < maxGenres; i++ {
		topGenres = append(topGenres, GenreStat{
			Genre: genres[i].genre,
			Count: genres[i].count,
		})
	}

	// Get top artists (sorted by count, descending)
	type artistCount struct {
		artistID string
		count    int
	}
	artists := make([]artistCount, 0, len(profile.ArtistListenCount))
	for artistID, count := range profile.ArtistListenCount {
		artists = append(artists, artistCount{artistID, count})
	}
	sort.Slice(artists, func(i, j int) bool {
		return artists[i].count > artists[j].count
	})

	topArtists := make([]ArtistStat, 0, 10)
	maxArtists := 10
	if maxArtists > len(artists) {
		maxArtists = len(artists)
	}
	for i := 0; i < maxArtists; i++ {
		topArtists = append(topArtists, ArtistStat{
			ArtistID: artists[i].artistID,
			Count:    artists[i].count,
		})
	}

	c.JSON(http.StatusOK, UserTasteResponse{
		UserID:     userID,
		TopGenres:  topGenres,
		TopArtists: topArtists,
	})
}

// TrendingQueriesResponse represents the trending queries API response
type TrendingQueriesResponse struct {
	Items []store.TrendingQuery `json:"items"`
}

func (s *Server) trendingQueriesHandler(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	days := 30 // Default: last 30 days
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	trending := s.searchQueryStore.GetTrendingQueries(limit, days)

	c.JSON(http.StatusOK, TrendingQueriesResponse{
		Items: trending,
	})
}

// RecordSearchQueryRequest represents the request to record a search query
type RecordSearchQueryRequest struct {
	Query string `json:"query" binding:"required"`
}

func (s *Server) recordSearchQueryHandler(c *gin.Context) {
	var req RecordSearchQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.searchQueryStore.Increment(req.Query)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// AddIngestEndpoints adds HTTP endpoints for manual event ingestion
func (s *Server) AddIngestEndpoints(
	trackStore store.TrackStore,
	userProfileStore store.UserProfileStore,
	globalStatsStore store.GlobalStatsStore,
) {
	s.router.POST("/ingest/track", func(c *gin.Context) {
		var req struct {
			EventType  string   `json:"event_type"`
			TrackID    string   `json:"track_id"`
			ArtistID   string   `json:"artist_id"`
			Genres     []string `json:"genres"`
			ReleaseTS  int64    `json:"release_ts"`
			IsExplicit bool     `json:"is_explicit"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		track := &models.Track{
			TrackID:    req.TrackID,
			ArtistID:   req.ArtistID,
			Genres:     req.Genres,
			ReleaseTS:  req.ReleaseTS,
			IsExplicit: req.IsExplicit,
		}
		trackStore.Upsert(track)

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	s.router.POST("/ingest/listening", func(c *gin.Context) {
		var req struct {
			EventType       string   `json:"event_type"`
			UserID          string   `json:"user_id"`
			TrackID         string   `json:"track_id"`
			ListenedSeconds int      `json:"listened_seconds"`
			Timestamp       int64    `json:"ts"`
			Lat             *float64 `json:"lat,omitempty"`
			Lon             *float64 `json:"lon,omitempty"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		profile := userProfileStore.GetOrCreate(req.UserID)

		track, ok := trackStore.Get(req.TrackID)
		if ok {
			for _, genre := range track.Genres {
				profile.GenreListenCount[genre]++
			}
			profile.ArtistListenCount[track.ArtistID]++
		}

		profile.ListenedTracks[req.TrackID] = struct{}{}
		userProfileStore.Update(profile)
		globalStatsStore.IncrementPlayCount(req.TrackID)

		// Process geo coordinates if enabled and provided
		if s.enableGeoTop && s.geoTopStore != nil && req.Lat != nil && req.Lon != nil {
			geohash := store.EncodeGeohash(*req.Lat, *req.Lon, s.geohashPrecision)
			s.geoTopStore.Incr(geohash, req.TrackID, 1)
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

// chartsTopNearbyHandler returns top tracks in a geographic radius
func (s *Server) chartsTopNearbyHandler(c *gin.Context) {
	// Check if feature is enabled
	if !s.enableGeoTop {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "Geo-based top charts feature is not enabled. Set ENABLE_GEO_TOP=true to enable.",
		})
		return
	}

	// Parse query parameters
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	radiusMStr := c.DefaultQuery("radius_m", strconv.Itoa(s.defaultRadiusM))
	limitStr := c.DefaultQuery("limit", "20")

	if latStr == "" || lonStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "lat and lon query parameters are required",
		})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid lat parameter",
		})
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid lon parameter",
		})
		return
	}

	radiusM, err := strconv.Atoi(radiusMStr)
	if err != nil || radiusM <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid radius_m parameter",
		})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid limit parameter",
		})
		return
	}

	// Validate coordinates
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid coordinates: lat must be in [-90, 90], lon in [-180, 180]",
		})
		return
	}

	// Determine precision based on radius
	precision := store.GeohashPrecisionForRadius(radiusM)

	// Get geohashes covering the radius
	geohashes := store.ExpandNeighbors(lat, lon, radiusM, precision)

	// Get top tracks for these geohashes
	topTracks := s.geoTopStore.GetTopForGeohashes(geohashes, limit)

	// Convert to response format
	trackIDs := make([]string, 0, len(topTracks))
	counts := make([]gin.H, 0, len(topTracks))
	for _, tc := range topTracks {
		trackIDs = append(trackIDs, tc.TrackID)
		counts = append(counts, gin.H{
			"track_id": tc.TrackID,
			"count":    tc.Count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"meta": gin.H{
			"precision":         precision,
			"radius_m":          radiusM,
			"geohashes_queried": len(geohashes),
			"center_lat":        lat,
			"center_lon":        lon,
		},
		"track_ids": trackIDs,
		"tracks":    counts,
	})
}
