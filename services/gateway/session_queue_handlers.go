package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	queuepb "github.com/MusicSocial/api-gateway/proto/queue/v1"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Session Queue Handlers - use playback_queue service with context_type="session"

// getSessionCurrentTrackHandler returns the current track in session's queue
func (g *Gateway) getSessionCurrentTrackHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user_id").(string)
	if !ok {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	roomId := vars["roomId"]
	if roomId == "" {
		writeError(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	resp, err := g.queueClient.GetCurrentTrack(ctx, &queuepb.GetCurrentTrackRequest{
		Context: &queuepb.ContextRef{
			ContextType: "session",
			ContextId:   roomId,
		},
	})

	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Queue is empty - this is normal
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"current": nil,
			})
			return
		}
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"current": resp.Current,
	})
}

// getSessionQueueHandler returns the list of tracks in session's queue
func (g *Gateway) getSessionQueueHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user_id").(string)
	if !ok {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	roomId := vars["roomId"]
	if roomId == "" {
		writeError(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	limit := int32(50)
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 32); err == nil {
			limit = int32(parsed)
		}
	}

	ctx := r.Context()
	resp, err := g.queueClient.ListQueue(ctx, &queuepb.ListQueueRequest{
		Context: &queuepb.ContextRef{
			ContextType: "session",
			ContextId:   roomId,
		},
		Limit: limit,
	})

	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Ensure items is always an array, not nil
	items := resp.Items
	if items == nil {
		items = []*queuepb.QueueItem{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": items,
	})
}

// addTrackToSessionQueueHandler adds a track to session's queue
func (g *Gateway) addTrackToSessionQueueHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user_id").(string)
	if !ok {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	roomId := vars["roomId"]
	if roomId == "" {
		writeError(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		TrackID string `json:"track_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.TrackID == "" {
		writeError(w, "track_id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	_, err := g.queueClient.EnqueueTrack(ctx, &queuepb.EnqueueTrackRequest{
		Context: &queuepb.ContextRef{
			ContextType: "session",
			ContextId:   roomId,
		},
		TrackId: req.TrackID,
	})

	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// removeTrackFromSessionQueueHandler removes a track from session's queue
func (g *Gateway) removeTrackFromSessionQueueHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user_id").(string)
	if !ok {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	roomId := vars["roomId"]
	trackID := vars["trackId"]
	if roomId == "" || trackID == "" {
		writeError(w, "Room ID and Track ID are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	resp, err := g.queueClient.RemoveTrack(ctx, &queuepb.RemoveTrackRequest{
		Context: &queuepb.ContextRef{
			ContextType: "session",
			ContextId:   roomId,
		},
		TrackId: trackID,
	})

	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": resp.Success,
	})
}

// getSessionNextTrackHandler gets the next track and moves cursor forward
func (g *Gateway) getSessionNextTrackHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user_id").(string)
	if !ok {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	roomId := vars["roomId"]
	if roomId == "" {
		writeError(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	resp, err := g.queueClient.GetNextTrack(ctx, &queuepb.GetNextTrackRequest{
		Context: &queuepb.ContextRef{
			ContextType: "session",
			ContextId:   roomId,
		},
	})

	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Queue is empty
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"next": nil,
			})
			return
		}
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"next": resp.Next,
	})
}

// getSessionPrevTrackHandler gets the previous track and moves cursor backward
func (g *Gateway) getSessionPrevTrackHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user_id").(string)
	if !ok {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	roomId := vars["roomId"]
	if roomId == "" {
		writeError(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	resp, err := g.queueClient.GetPrevTrack(ctx, &queuepb.GetPrevTrackRequest{
		Context: &queuepb.ContextRef{
			ContextType: "session",
			ContextId:   roomId,
		},
	})

	if err != nil {
		if status.Code(err) == codes.NotFound {
			// No previous track
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"previous": nil,
			})
			return
		}
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"previous": resp.Previous,
	})
}

// clearSessionQueueHandler clears session's queue
func (g *Gateway) clearSessionQueueHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user_id").(string)
	if !ok {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	roomId := vars["roomId"]
	if roomId == "" {
		writeError(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	_, err := g.queueClient.ClearQueue(ctx, &queuepb.ClearQueueRequest{
		Context: &queuepb.ContextRef{
			ContextType: "session",
			ContextId:   roomId,
		},
	})

	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

