package api

import (
	"net/http"

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
	router           *gin.Engine
}

// NewServer creates a new HTTP server
func NewServer(
	port string,
	recEngine *engine.RecommendationEngine,
	userProfileStore store.UserProfileStore,
	trackStore store.TrackStore,
) *Server {
	s := &Server{
		port:             port,
		engine:           recEngine,
		userProfileStore: userProfileStore,
		trackStore:       trackStore,
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
			EventType       string `json:"event_type"`
			UserID          string `json:"user_id"`
			TrackID         string `json:"track_id"`
			ListenedSeconds int    `json:"listened_seconds"`
			Timestamp       int64  `json:"ts"`
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

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
