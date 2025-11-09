package internal

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Public API
	mux.HandleFunc("/api/tracks", h.handleTracks)
	mux.HandleFunc("/api/tracks/", h.handleTrack)

	// Admin API
	mux.HandleFunc("/api/admin/tracks", h.handleAdminTracks)
	mux.HandleFunc("/api/admin/tracks/", h.handleAdminTrack)

	// Health
	mux.HandleFunc("/health", h.health)
}

// GET /api/tracks - список треков
// GET /api/tracks?artist_id=uuid&limit=20&offset=0
func (h *Handler) handleTracks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	var artistID *uuid.UUID
	if aid := r.URL.Query().Get("artist_id"); aid != "" {
		if id, err := uuid.Parse(aid); err == nil {
			artistID = &id
		}
	}

	tracks, err := h.service.ListTracks(r.Context(), limit, offset, artistID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"tracks": tracks,
		"limit":  limit,
		"offset": offset,
	})
}

// GET /api/tracks/:id - один трек
func (h *Handler) handleTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID из пути /api/tracks/{id}
	path := r.URL.Path[len("/api/tracks/"):]
	id, err := uuid.Parse(path)
	if err != nil {
		http.Error(w, "Invalid track ID", http.StatusBadRequest)
		return
	}

	track, err := h.service.GetTrack(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		http.Error(w, "Track not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, track)
}

// POST /api/admin/tracks - создать трек
func (h *Handler) handleAdminTracks(w http.ResponseWriter, r *http.Request) {
	// Простая проверка роли через header
	if r.Header.Get("X-User-Role") != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Title     string   `json:"title"`
		ArtistIDs []string `json:"artist_ids"` // Массив UUID артистов
		Genre     string   `json:"genre"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.ArtistIDs) == 0 {
		http.Error(w, "At least one artist_id is required", http.StatusBadRequest)
		return
	}

	// Парсим UUID артистов
	artistIDs, err := parseUUIDs(req.ArtistIDs)
	if err != nil {
		http.Error(w, "Invalid artist ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	track, err := h.service.CreateTrack(r.Context(), req.Title, artistIDs, req.Genre)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, track)
}

// PUT /api/admin/tracks/:id - обновить трек
// DELETE /api/admin/tracks/:id - удалить трек
func (h *Handler) handleAdminTrack(w http.ResponseWriter, r *http.Request) {
	// Проверка роли
	if r.Header.Get("X-User-Role") != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path[len("/api/admin/tracks/"):]
	id, err := uuid.Parse(path)
	if err != nil {
		http.Error(w, "Invalid track ID", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPut:
		var req struct {
			Title     string   `json:"title"`
			ArtistIDs []string `json:"artist_ids"` // Массив UUID артистов (опционально)
			Genre     string   `json:"genre"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Парсим UUID артистов, если они переданы
		var artistIDs []uuid.UUID
		if len(req.ArtistIDs) > 0 {
			var err error
			artistIDs, err = parseUUIDs(req.ArtistIDs)
			if err != nil {
				http.Error(w, "Invalid artist ID: "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		if err := h.service.UpdateTrack(r.Context(), id, req.Title, artistIDs, req.Genre); err != nil {
			if errors.Is(err, ErrNotFound) {
				http.Error(w, "Track not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		track, _ := h.service.GetTrack(r.Context(), id)
		respondJSON(w, http.StatusOK, track)

	case http.MethodDelete:
		if err := h.service.DeleteTrack(r.Context(), id); err != nil {
			if errors.Is(err, ErrNotFound) {
				http.Error(w, "Track not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
