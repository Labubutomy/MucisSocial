package internal

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует все маршруты приложения
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", h.health)

	// API routes
	api := router.Group("/api")
	{
		h.registerPlaylistRoutes(api)
		h.registerUserPlaylistRoutes(api)
	}
}

// регистрирует маршруты для работы с плейлистами
func (h *Handler) registerPlaylistRoutes(api *gin.RouterGroup) {
	playlists := api.Group("/playlists")
	{
		// Создание плейлиста
		playlists.POST("", h.createPlaylist)

		// Операции с конкретным плейлистом
		playlist := playlists.Group("/:id")
		{
			playlist.DELETE("", h.deletePlaylist)
			playlist.GET("/tracks", h.getPlaylistTracks)
			playlist.GET("/author", h.getPlaylistAuthor)

			// Операции с треками в плейлисте
			playlist.DELETE("/tracks/:track_id", h.removeTrackFromPlaylist)
		}
	}
}

// регистрирует маршруты для работы с подписками пользователей
func (h *Handler) registerUserPlaylistRoutes(api *gin.RouterGroup) {
	users := api.Group("/users")
	{
		user := users.Group("/:user_id")
		{
			playlists := user.Group("/playlists")
			{
				// Список подписанных плейлистов
				playlists.GET("", h.getUserPlaylists)

				// Операции с подпиской на плейлист
				playlist := playlists.Group("/:playlist_id")
				{
					playlist.POST("/subscribe", h.subscribeUser)
					playlist.DELETE("/subscribe", h.unsubscribeUser)
				}
			}
		}
	}
}

type CreatePlaylistRequest struct {
	AuthorID uuid.UUID `json:"author_id" binding:"required"`
	Name     string    `json:"name" binding:"required"`
	TrackIDs []string  `json:"track_ids"`
}

func parseUUIDParam(c *gin.Context, paramName string) (uuid.UUID, error) {
	idStr := c.Param(paramName)
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid " + paramName + ": " + idStr})
		return uuid.Nil, err
	}
	return id, nil
}

func handleError(c *gin.Context, err error) {
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

// создает новый плейлист
// POST /api/playlists
func (h *Handler) createPlaylist(c *gin.Context) {
	var req CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Парсим track IDs
	trackIDs, err := parseTrackIDs(req.TrackIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	playlist, err := h.service.CreatePlaylist(c.Request.Context(), req.AuthorID, req.Name, trackIDs)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, playlist)
}

// удаляет плейлист
// DELETE /api/playlists/:id
func (h *Handler) deletePlaylist(c *gin.Context) {
	playlistID, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}

	if err := h.service.DeletePlaylist(c.Request.Context(), playlistID); err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// возвращает список треков из плейлиста
// GET /api/playlists/:id/tracks
func (h *Handler) getPlaylistTracks(c *gin.Context) {
	playlistID, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}

	tracks, err := h.service.GetPlaylistTracks(c.Request.Context(), playlistID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"playlist_id": playlistID,
		"tracks":      tracks,
	})
}

// возвращает автора плейлиста
// GET /api/playlists/:id/author
func (h *Handler) getPlaylistAuthor(c *gin.Context) {
	playlistID, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}

	authorID, err := h.service.GetPlaylistAuthor(c.Request.Context(), playlistID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"playlist_id": playlistID,
		"author_id":   authorID,
	})
}

// удаляет трек из плейлиста
// DELETE /api/playlists/:id/tracks/:track_id
func (h *Handler) removeTrackFromPlaylist(c *gin.Context) {
	playlistID, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}

	trackID, err := parseUUIDParam(c, "track_id")
	if err != nil {
		return
	}

	if err := h.service.RemoveTrackFromPlaylist(c.Request.Context(), playlistID, trackID); err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// возвращает список подписанных плейлистов пользователя
// GET /api/users/:user_id/playlists?limit=20&offset=0
func (h *Handler) getUserPlaylists(c *gin.Context) {
	userID, err := parseUUIDParam(c, "user_id")
	if err != nil {
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	playlists, err := h.service.GetUserPlaylists(c.Request.Context(), userID, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"playlists": playlists,
		"limit":     limit,
		"offset":    offset,
	})
}

// подписывает пользователя на плейлист
// POST /api/users/:user_id/playlists/:playlist_id/subscribe
func (h *Handler) subscribeUser(c *gin.Context) {
	userID, err := parseUUIDParam(c, "user_id")
	if err != nil {
		return
	}

	playlistID, err := parseUUIDParam(c, "playlist_id")
	if err != nil {
		return
	}

	if err := h.service.SubscribeUser(c.Request.Context(), userID, playlistID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "subscribed"})
}

// отписывает пользователя от плейлиста
// DELETE /api/users/:user_id/playlists/:playlist_id/subscribe
func (h *Handler) unsubscribeUser(c *gin.Context) {
	userID, err := parseUUIDParam(c, "user_id")
	if err != nil {
		return
	}

	playlistID, err := parseUUIDParam(c, "playlist_id")
	if err != nil {
		return
	}

	if err := h.service.UnsubscribeUser(c.Request.Context(), userID, playlistID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "unsubscribed"})
}

// проверяет состояние сервиса
// GET /health
func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// парсит массив строк в массив UUID
func parseTrackIDs(trackIDStrings []string) ([]uuid.UUID, error) {
	if len(trackIDStrings) == 0 {
		return nil, nil
	}

	trackIDs := make([]uuid.UUID, 0, len(trackIDStrings))
	for _, trackIDStr := range trackIDStrings {
		trackID, err := uuid.Parse(trackIDStr)
		if err != nil {
			return nil, errors.New("Invalid track ID: " + trackIDStr)
		}
		trackIDs = append(trackIDs, trackID)
	}
	return trackIDs, nil
}
