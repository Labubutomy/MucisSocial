package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	sessionpb "github.com/MusicSocial/api-gateway/proto/session/v1"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RoomStateJSON represents the JSON format expected by frontend
type RoomStateJSON struct {
	RoomID      string                `json:"roomId"`
	CurrentTrack *TrackInfoJSON       `json:"currentTrack,omitempty"`
	Position    float64               `json:"position"`
	IsPlaying   bool                  `json:"isPlaying"`
	Participants []ParticipantJSON    `json:"participants"`
	Queue       []TrackInfoJSON       `json:"queue"`
	LastAction  *ActionJSON           `json:"lastAction,omitempty"`
	CreatedAt   string                `json:"createdAt"`
	UpdatedAt   string                `json:"updatedAt"`
}

type TrackInfoJSON struct {
	TrackID  string  `json:"trackId"`
	Title    string  `json:"title"`
	Artist   string  `json:"artist"`
	Duration float64 `json:"duration"`
	CDNURL   string  `json:"cdnUrl"`
}

type ParticipantJSON struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	IsOnline bool   `json:"isOnline"`
	JoinedAt string `json:"joinedAt"`
}

type ActionJSON struct {
	ActionID  string            `json:"actionId"`
	Type      string            `json:"type"`
	UserID    string            `json:"userId"`
	Timestamp string            `json:"timestamp"`
	Payload   map[string]string `json:"payload,omitempty"`
}

// convertProtoToJSON converts gRPC RoomState to JSON format
func convertProtoToJSON(proto *sessionpb.RoomState) *RoomStateJSON {
	result := &RoomStateJSON{
		RoomID:    proto.RoomId,
		Position:  proto.Position,
		IsPlaying: proto.IsPlaying,
		Queue:     make([]TrackInfoJSON, 0, len(proto.Queue)),
		Participants: make([]ParticipantJSON, 0, len(proto.Participants)),
	}

	if proto.CurrentTrack != nil {
		result.CurrentTrack = &TrackInfoJSON{
			TrackID:  proto.CurrentTrack.TrackId,
			Title:    proto.CurrentTrack.Title,
			Artist:   proto.CurrentTrack.Artist,
			Duration: proto.CurrentTrack.Duration,
			CDNURL:   proto.CurrentTrack.CdnUrl,
		}
	}

	for _, track := range proto.Queue {
		result.Queue = append(result.Queue, TrackInfoJSON{
			TrackID:  track.TrackId,
			Title:    track.Title,
			Artist:   track.Artist,
			Duration: track.Duration,
			CDNURL:   track.CdnUrl,
		})
	}

	for _, participant := range proto.Participants {
		joinedAt := ""
		if participant.JoinedAt != nil {
			joinedAt = timestampToISO(participant.JoinedAt)
		}
		result.Participants = append(result.Participants, ParticipantJSON{
			UserID:   participant.UserId,
			Username: participant.Username,
			IsOnline: participant.IsOnline,
			JoinedAt: joinedAt,
		})
	}

	if proto.LastAction != nil {
		actionTime := ""
		if proto.LastAction.Timestamp != nil {
			actionTime = timestampToISO(proto.LastAction.Timestamp)
		}
		result.LastAction = &ActionJSON{
			ActionID:  proto.LastAction.ActionId,
			Type:      proto.LastAction.Type.String(),
			UserID:    proto.LastAction.UserId,
			Timestamp: actionTime,
			Payload:   proto.LastAction.Payload,
		}
	}

	if proto.CreatedAt != nil {
		result.CreatedAt = timestampToISO(proto.CreatedAt)
	}
	if proto.UpdatedAt != nil {
		result.UpdatedAt = timestampToISO(proto.UpdatedAt)
	}

	return result
}

// timestampToISO converts protobuf Timestamp to ISO 8601 string
func timestampToISO(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().Format(time.RFC3339Nano)
}

// sessionHealthHandler godoc
//
//	@Summary		Проверка состояния session service
//	@Description	Проксирует health check запрос на session service через gRPC
//	@Tags			Session
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/api/rooms/health [get]
func (g *Gateway) sessionHealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := g.sessionClient.HealthCheck(ctx, &sessionpb.HealthCheckRequest{})
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": resp.Status,
	})
}

// getRoomHandler godoc
//
//	@Summary		Получить комнату
//	@Description	Получает состояние комнаты по ID через gRPC
//	@Tags			Session
//	@Security		BearerAuth
//	@Param			roomId	path		string	true	"Room ID"
//	@Success		200		{object}	object
//	@Failure		404		{object}	ErrorResponse
//	@Router			/api/v1/rooms/{roomId} [get]
func (g *Gateway) getRoomHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomId := vars["roomId"]
	if roomId == "" {
		writeError(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := g.sessionClient.GetRoom(ctx, &sessionpb.GetRoomRequest{RoomId: roomId})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			writeError(w, "Room not found", http.StatusNotFound)
			return
		}
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonRoom := convertProtoToJSON(resp.Room)
	json.NewEncoder(w).Encode(jsonRoom)
}

// createRoomHandler godoc
//
//	@Summary		Создать комнату
//	@Description	Создает новую комнату совместного прослушивания через gRPC
//	@Tags			Session
//	@Security		BearerAuth
//	@Param			roomId	path		string	true	"Room ID"
//	@Success		200		{object}	object
//	@Router			/api/v1/rooms/{roomId} [post]
func (g *Gateway) createRoomHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomId := vars["roomId"]
	if roomId == "" {
		writeError(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := g.sessionClient.CreateRoom(ctx, &sessionpb.CreateRoomRequest{RoomId: roomId})
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonRoom := convertProtoToJSON(resp.Room)
	json.NewEncoder(w).Encode(jsonRoom)
}

// deleteRoomHandler godoc
//
//	@Summary		Удалить комнату
//	@Description	Удаляет комнату совместного прослушивания через gRPC
//	@Tags			Session
//	@Security		BearerAuth
//	@Param			roomId	path		string	true	"Room ID"
//	@Success		204
//	@Router			/api/v1/rooms/{roomId} [delete]
func (g *Gateway) deleteRoomHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomId := vars["roomId"]
	if roomId == "" {
		writeError(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := g.sessionClient.DeleteRoom(ctx, &sessionpb.DeleteRoomRequest{RoomId: roomId})
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// addParticipantHandler godoc
//
//	@Summary		Добавить участника
//	@Description	Добавляет участника в комнату через gRPC
//	@Tags			Session
//	@Security		BearerAuth
//	@Param			roomId	path		string	true	"Room ID"
//	@Param			userId	query		string	true	"User ID"
//	@Param			username	query		string	true	"Username"
//	@Success		200		{object}	object
//	@Router			/api/v1/rooms/{roomId}/participants [post]
func (g *Gateway) addParticipantHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomId := vars["roomId"]
	userId := r.URL.Query().Get("userId")
	username := r.URL.Query().Get("username")

	if roomId == "" || userId == "" || username == "" {
		writeError(w, "Room ID, User ID and Username are required", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(string)
	if userID != userId {
		writeError(w, "Cannot add other user", http.StatusForbidden)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := g.sessionClient.AddParticipant(ctx, &sessionpb.AddParticipantRequest{
		RoomId:   roomId,
		UserId:   userId,
		Username: username,
	})
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonRoom := convertProtoToJSON(resp.Room)
	json.NewEncoder(w).Encode(jsonRoom)
}

// removeParticipantHandler godoc
//
//	@Summary		Удалить участника
//	@Description	Удаляет участника из комнаты через gRPC
//	@Tags			Session
//	@Security		BearerAuth
//	@Param			roomId	path		string	true	"Room ID"
//	@Param			userId	path		string	true	"User ID"
//	@Success		200		{object}	object
//	@Router			/api/v1/rooms/{roomId}/participants/{userId} [delete]
func (g *Gateway) removeParticipantHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomId := vars["roomId"]
	userId := vars["userId"]

	if roomId == "" || userId == "" {
		writeError(w, "Room ID and User ID are required", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(string)
	if userID != userId {
		writeError(w, "Cannot remove other user", http.StatusForbidden)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := g.sessionClient.RemoveParticipant(ctx, &sessionpb.RemoveParticipantRequest{
		RoomId: roomId,
		UserId: userId,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			writeError(w, "Room or participant not found", http.StatusNotFound)
			return
		}
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonRoom := convertProtoToJSON(resp.Room)
	json.NewEncoder(w).Encode(jsonRoom)
}


