// Package main MucissSocial API Gateway
//
//	@title			MucissSocial API Gateway
//	@version		1.0.0
//	@description	API Gateway для сервиса MucissSocial. Предоставляет REST API для взаимодействия с микросервисами.
//
//	@contact.name	MucissSocial Team
//
//	@host		localhost:8080
//	@BasePath	/
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				JWT токен в формате 'Bearer {token}'
//
//	@schemes	http https
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	artistpb "github.com/MusicSocial/api-gateway/proto/artists/v1"
	playlistpb "github.com/MusicSocial/api-gateway/proto/playlist/v1"
	trackspb "github.com/MusicSocial/api-gateway/proto/tracks/v1"
	uploadpb "github.com/MusicSocial/api-gateway/proto/upload"
	pb "github.com/MusicSocial/api-gateway/proto/users/v1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	_ "github.com/MusicSocial/api-gateway/docs"
)

type Gateway struct {
	userClient     pb.UserServiceClient
	artistClient   artistpb.ArtistServiceClient
	tracksClient   trackspb.TracksServiceClient
	playlistClient playlistpb.PlaylistServiceClient
	uploadClient   uploadpb.UploadServiceClient
	jwtSecret      []byte
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	// Connect to user gRPC service
	userConn, err := grpc.Dial("users-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to users gRPC service: %v", err)
	}
	defer userConn.Close()

	// Connect to artist gRPC service
	artistConn, err := grpc.Dial("artists-service:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to artists gRPC service: %v", err)
	}
	defer artistConn.Close()

	// Connect to tracks gRPC service
	tracksConn, err := grpc.Dial("tracks-service:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to tracks gRPC service: %v", err)
	}
	defer tracksConn.Close()

	// Connect to playlist gRPC service
	playlistConn, err := grpc.Dial("playlist-service:50054", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to playlist gRPC service: %v", err)
	}
	defer playlistConn.Close()

	// Connect to upload gRPC service
	uploadConn, err := grpc.Dial("upload-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to upload gRPC service: %v", err)
	}
	defer uploadConn.Close()

	gateway := &Gateway{
		userClient:     pb.NewUserServiceClient(userConn),
		artistClient:   artistpb.NewArtistServiceClient(artistConn),
		tracksClient:   trackspb.NewTracksServiceClient(tracksConn),
		playlistClient: playlistpb.NewPlaylistServiceClient(playlistConn),
		uploadClient:   uploadpb.NewUploadServiceClient(uploadConn),
		jwtSecret:      []byte(getEnv("JWT_SECRET", "your-super-secret-access-key-change-in-production")),
	}

	r := mux.NewRouter()

	// CORS middleware
	r.Use(corsMiddleware)

	// Swagger documentation
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Health check
	r.HandleFunc("/__health", healthHandler).Methods("GET")

	// Auth endpoints (no JWT required)
	r.HandleFunc("/api/v1/auth/sign-up", gateway.signUpHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/auth/sign-in", gateway.signInHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/auth/refresh", gateway.refreshHandler).Methods("POST", "OPTIONS")

	// Public artist endpoints (no JWT required) - specific routes first
	r.HandleFunc("/api/v1/artists/trending", gateway.getTrendingArtistsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/artists/search", gateway.searchArtistsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/artists/{artistId}", gateway.getArtistByIdHandler).Methods("GET", "OPTIONS")

	// Public tracks endpoints (no JWT required)
	r.HandleFunc("/api/v1/tracks", gateway.getTracksHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/tracks/search", gateway.searchTracksHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/tracks/{trackId}", gateway.getTrackByIdHandler).Methods("GET", "OPTIONS")

	// Upload endpoint (no JWT required)
	r.HandleFunc("/api/v1/upload/track", gateway.uploadTrackHandler).Methods("POST", "OPTIONS")

	// Protected endpoints (JWT required)
	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Use(gateway.jwtMiddleware)
	protected.HandleFunc("/me", gateway.getMeHandler).Methods("GET", "OPTIONS")
	protected.HandleFunc("/me", gateway.updateMeHandler).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/me/search-history", gateway.getSearchHistoryHandler).Methods("GET", "OPTIONS")
	protected.HandleFunc("/me/search-history", gateway.addSearchHistoryHandler).Methods("POST", "OPTIONS")
	protected.HandleFunc("/me/search-history", gateway.clearSearchHistoryHandler).Methods("DELETE", "OPTIONS")

	// Artists endpoints
	protected.HandleFunc("/artists", gateway.createArtistHandler).Methods("POST", "OPTIONS")

	// Tracks endpoints
	protected.HandleFunc("/tracks", gateway.createTrackHandler).Methods("POST", "OPTIONS")
	protected.HandleFunc("/tracks/{trackId}", gateway.updateTrackInfoHandler).Methods("PUT", "OPTIONS")

	// Playlist endpoints
	protected.HandleFunc("/playlists", gateway.createPlaylistHandler).Methods("POST", "OPTIONS")
	protected.HandleFunc("/playlists", gateway.getUserPlaylistsHandler).Methods("GET", "OPTIONS")
	protected.HandleFunc("/playlists/{playlistId}", gateway.getPlaylistHandler).Methods("GET", "OPTIONS")
	protected.HandleFunc("/playlists/{playlistId}", gateway.updatePlaylistHandler).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/playlists/{playlistId}", gateway.deletePlaylistHandler).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/playlists/{playlistId}/tracks", gateway.getPlaylistTracksHandler).Methods("GET", "OPTIONS")
	protected.HandleFunc("/playlists/{playlistId}/tracks", gateway.addTrackToPlaylistHandler).Methods("POST", "OPTIONS")
	protected.HandleFunc("/playlists/{playlistId}/tracks/{trackId}", gateway.removeTrackFromPlaylistHandler).Methods("DELETE", "OPTIONS")

	port := getEnv("PORT", "8080")
	log.Printf("API Gateway starting on port %s", port)
	log.Printf("Swagger documentation available at: http://localhost:%s/swagger/", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (g *Gateway) jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		var tokenString string
		
		// Поддерживаем оба формата: "Bearer <token>" и "Bearer: <token>"
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else if strings.HasPrefix(authHeader, "Bearer:") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer:")
			tokenString = strings.TrimSpace(tokenString)
		} else {
			writeError(w, "Invalid Authorization header format. Expected 'Bearer <token>' or 'Bearer: <token>'", http.StatusUnauthorized)
			return
		}

		if tokenString == "" {
			writeError(w, "Token is required", http.StatusUnauthorized)
			return
		}
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return g.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			writeError(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			ctx := context.WithValue(r.Context(), "user_id", claims["user_id"])
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			writeError(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
	})
}

// healthHandler godoc
//
//	@Summary		Проверка состояния сервиса
//	@Description	Возвращает статус работоспособности API Gateway
//	@Tags			Health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/__health [get]
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// signUpHandler godoc
//
//	@Summary		Регистрация нового пользователя
//	@Description	Создание нового аккаунта пользователя
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{email=string,password=string,username=string}	true	"Данные для регистрации"
//	@Success		200		{object}	object{access_token=string,refresh_token=string,user=object}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Router			/api/v1/auth/sign-up [post]
func (g *Gateway) signUpHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	grpcReq := &pb.SignUpRequest{
		Email:    req.Email,
		Password: req.Password,
		Username: req.Username,
	}

	resp, err := g.userClient.SignUp(context.Background(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user":          resp.User,
	})
}

// signInHandler godoc
//
//	@Summary		Вход в систему
//	@Description	Аутентификация пользователя по email и паролю
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{email=string,password=string}	true	"Данные для входа"
//	@Success		200		{object}	object{access_token=string,refresh_token=string,user=object}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/api/v1/auth/sign-in [post]
func (g *Gateway) signInHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	grpcReq := &pb.SignInRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := g.userClient.SignIn(context.Background(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user":          resp.User,
	})
}

// refreshHandler godoc
//
//	@Summary		Обновление токена
//	@Description	Получение нового access token с помощью refresh token
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{refresh_token=string}	true	"Refresh token"
//	@Success		200		{object}	object{access_token=string,refresh_token=string}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/api/v1/auth/refresh [post]
func (g *Gateway) refreshHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	grpcReq := &pb.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	resp, err := g.userClient.RefreshToken(context.Background(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

// getMeHandler godoc
//
//	@Summary		Получение профиля текущего пользователя
//	@Description	Возвращает информацию о текущем аутентифицированном пользователе
//	@Tags			User Profile
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	object{user=object}
//	@Failure		401	{object}	ErrorResponse
//	@Router			/api/v1/me [get]
func (g *Gateway) getMeHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	grpcReq := &pb.GetMeRequest{
		UserId: userID,
	}

	resp, err := g.userClient.GetMe(context.Background(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": resp.User,
	})
}

// updateMeHandler godoc
//
//	@Summary		Обновление профиля пользователя
//	@Description	Обновление информации профиля текущего пользователя
//	@Tags			User Profile
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		object{username=string,avatar_url=string}	true	"Данные для обновления профиля"
//	@Success		200		{object}	object{user=object}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/api/v1/me [put]
func (g *Gateway) updateMeHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req struct {
		Username  string `json:"username"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	grpcReq := &pb.UpdateProfileRequest{
		UserId:    userID,
		Username:  &req.Username,
		AvatarUrl: &req.AvatarURL,
	}

	resp, err := g.userClient.UpdateProfile(context.Background(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": resp.User,
	})
}

// getSearchHistoryHandler godoc
//
//	@Summary		Получение истории поиска
//	@Description	Возвращает историю поиска пользователя
//	@Tags			Search History
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		int	false	"Количество записей для возврата (по умолчанию 10)"	default(10)
//	@Success		200		{object}	object{items=[]object}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/api/v1/me/search-history [get]
func (g *Gateway) getSearchHistoryHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	limit := r.URL.Query().Get("limit")

	var limitInt int32 = 10
	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			limitInt = int32(l)
		} else {
			writeError(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
	}

	grpcReq := &pb.GetSearchHistoryRequest{
		UserId: userID,
		Limit:  limitInt,
	}

	resp, err := g.userClient.GetSearchHistory(context.Background(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": resp.Items,
	})
}

// addSearchHistoryHandler godoc
//
//	@Summary		Добавление записи в историю поиска
//	@Description	Добавляет новый поисковый запрос в историю пользователя
//	@Tags			Search History
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		object{query=string}	true	"Поисковый запрос"
//	@Success		200		{object}	object{item=object}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/api/v1/me/search-history [post]
func (g *Gateway) addSearchHistoryHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	grpcReq := &pb.AddSearchHistoryRequest{
		UserId: userID,
		Query:  req.Query,
	}

	resp, err := g.userClient.AddSearchHistory(context.Background(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"item": resp.Item,
	})
}

// clearSearchHistoryHandler godoc
//
//	@Summary		Очистка истории поиска
//	@Description	Удаляет всю историю поиска пользователя
//	@Tags			Search History
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	object{success=bool}
//	@Failure		401	{object}	ErrorResponse
//	@Router			/api/v1/me/search-history [delete]
func (g *Gateway) clearSearchHistoryHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	grpcReq := &pb.ClearSearchHistoryRequest{
		UserId: userID,
	}

	resp, err := g.userClient.ClearSearchHistory(context.Background(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": resp.Success,
	})
}

func writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   message,
		Code:    code,
		Message: message,
	})
}

func handleGrpcError(w http.ResponseWriter, err error) {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.NotFound:
			writeError(w, st.Message(), http.StatusNotFound)
		case codes.InvalidArgument:
			writeError(w, st.Message(), http.StatusBadRequest)
		case codes.Unauthenticated:
			writeError(w, st.Message(), http.StatusUnauthorized)
		case codes.PermissionDenied:
			writeError(w, st.Message(), http.StatusForbidden)
		case codes.AlreadyExists:
			writeError(w, st.Message(), http.StatusConflict)
		default:
			writeError(w, st.Message(), http.StatusInternalServerError)
		}
	} else {
		writeError(w, "Internal server error", http.StatusInternalServerError)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Artist Handlers

// getArtistByIdHandler godoc
//
//	@Summary		Get artist by ID
//	@Description	Get detailed information about a specific artist
//	@Tags			Artists
//	@Accept			json
//	@Produce		json
//	@Param			artistId	path		string	true	"Artist ID"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Router			/api/v1/artists/{artistId} [get]
func (g *Gateway) getArtistByIdHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	artistId := vars["artistId"]

	if artistId == "" {
		writeError(w, "Artist ID is required", http.StatusBadRequest)
		return
	}

	req := &artistpb.GetArtistByIdRequest{
		Id: artistId,
	}

	resp, err := g.artistClient.GetArtistById(r.Context(), req)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Artist)
}

// getTrendingArtistsHandler godoc
//
//	@Summary		Get trending artists
//	@Description	Get list of trending artists
//	@Tags			Artists
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/v1/artists/trending [get]
func (g *Gateway) getTrendingArtistsHandler(w http.ResponseWriter, r *http.Request) {
	req := &artistpb.GetTrendingArtistsRequest{
		Limit: 20, // Default limit
	}

	resp, err := g.artistClient.GetTrendingArtists(r.Context(), req)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	result := map[string]interface{}{
		"items": resp.Artists,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// searchArtistsHandler godoc
//
//	@Summary		Search artists
//	@Description	Search for artists by query
//	@Tags			Artists
//	@Accept			json
//	@Produce		json
//	@Param			q	query		string	true	"Search query"
//	@Success		200	{object}	map[string]interface{}
//	@Failure		400	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/v1/artists/search [get]
func (g *Gateway) searchArtistsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, "Search query is required", http.StatusBadRequest)
		return
	}

	req := &artistpb.SearchArtistsRequest{
		Query: query,
		Limit: 20, // Default limit
	}

	resp, err := g.artistClient.SearchArtists(r.Context(), req)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	// Ensure items is always an array, not nil
	items := resp.Artists
	if items == nil {
		items = []*artistpb.ArtistResponse{}
	}

	result := map[string]interface{}{
		"query": query,
		"items": items,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// createArtistHandler godoc
//
//	@Summary		Создать артиста
//	@Description	Создание нового артиста
//	@Tags			Artists
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		object{name=string,avatar_url=string,genres=[]string}	true	"Данные для создания артиста"
//	@Success		201		{object}	object{artist=object}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/v1/artists [post]
func (g *Gateway) createArtistHandler(w http.ResponseWriter, r *http.Request) {
	// Дефолтная обложка для всех артистов
	const defaultAvatarURL = "https://mir-s3-cdn-cf.behance.net/projects/202/e2ba0e187042211.Y3JvcCw4MDgsNjMyLDAsMA.png"

	var req struct {
		Name     string   `json:"name"`
		AvatarURL string  `json:"avatar_url"`
		Genres   []string `json:"genres"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		writeError(w, "name is required", http.StatusBadRequest)
		return
	}

	// Используем дефолтную обложку, если не указана
	avatarURL := req.AvatarURL
	if avatarURL == "" {
		avatarURL = defaultAvatarURL
	}

	grpcReq := &artistpb.CreateArtistRequest{
		Name:     req.Name,
		AvatarUrl: avatarURL,
		Genres:   req.Genres,
	}

	resp, err := g.artistClient.CreateArtist(r.Context(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"artist": resp.Artist,
	})
}

// ===== TRACKS HANDLERS =====

// @Summary Создать трек
// @Description Создание нового трека
// @Tags tracks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateTrackRequest true "Данные для создания трека"
// @Success 201 {object} CreateTrackResponse "Трек успешно создан"
// @Failure 400 {object} ErrorResponse "Некорректные данные запроса"
// @Failure 401 {object} ErrorResponse "Пользователь не авторизован"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/tracks [post]
func (g *Gateway) createTrackHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title     string   `json:"title"`
		ArtistIds []string `json:"artist_ids"`
		Genre     string   `json:"genre"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	grpcReq := &trackspb.CreateTrackRequest{
		Title:     req.Title,
		ArtistIds: req.ArtistIds,
		Genre:     req.Genre,
	}

	resp, err := g.tracksClient.CreateTrack(r.Context(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	result := map[string]interface{}{
		"track_id": resp.TrackId,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// @Summary Обновить информацию о треке
// @Description Обновление информации о треке (cover_url, audio_url, duration)
// @Tags tracks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param trackId path string true "ID трека"
// @Param request body UpdateTrackInfoRequest true "Данные для обновления трека"
// @Success 200 {object} map[string]bool "Трек успешно обновлен"
// @Failure 400 {object} ErrorResponse "Некорректные данные запроса"
// @Failure 401 {object} ErrorResponse "Пользователь не авторизован"
// @Failure 404 {object} ErrorResponse "Трек не найден"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/tracks/{trackId} [put]
func (g *Gateway) updateTrackInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trackId := vars["trackId"]

	var req struct {
		CoverUrl    string `json:"cover_url"`
		AudioUrl    string `json:"audio_url"`
		DurationSec int32  `json:"duration_sec"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	grpcReq := &trackspb.UpdateTrackInfoRequest{
		TrackId:     trackId,
		CoverUrl:    req.CoverUrl,
		AudioUrl:    req.AudioUrl,
		DurationSec: req.DurationSec,
	}

	resp, err := g.tracksClient.UpdateTrackInfo(r.Context(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	result := map[string]bool{
		"success": resp.Success,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ===== PLAYLIST HANDLERS =====

// @Summary Создать плейлист
// @Description Создание нового плейлиста
// @Tags playlists
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreatePlaylistRequest true "Данные для создания плейлиста"
// @Success 201 {object} CreatePlaylistResponse "Плейлист успешно создан"
// @Failure 400 {object} ErrorResponse "Некорректные данные запроса"
// @Failure 401 {object} ErrorResponse "Пользователь не авторизован"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/playlists [post]
func (g *Gateway) createPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user_id").(string)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsPrivate   bool   `json:"is_private"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	grpcReq := &playlistpb.CreatePlaylistRequest{
		UserId:      userId,
		Name:        req.Name,
		Description: req.Description,
		IsPrivate:   req.IsPrivate,
	}

	resp, err := g.playlistClient.CreatePlaylist(r.Context(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	result := map[string]interface{}{
		"playlist_id": resp.PlaylistId,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// @Summary Получить плейлисты пользователя
// @Description Получение списка плейлистов текущего пользователя
// @Tags playlists
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Количество записей на странице" default(20)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {object} GetUserPlaylistsResponse "Список плейлистов"
// @Failure 401 {object} ErrorResponse "Пользователь не авторизован"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/playlists [get]
func (g *Gateway) getUserPlaylistsHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user_id").(string)

	limit := 20
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	grpcReq := &playlistpb.GetUserPlaylistsRequest{
		UserId: userId,
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	resp, err := g.playlistClient.GetUserPlaylists(r.Context(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	// Убеждаемся, что playlists не nil (если nil, возвращаем пустой массив)
	playlists := resp.Playlists
	if playlists == nil {
		playlists = []*playlistpb.Playlist{}
	}

	result := map[string]interface{}{
		"playlists": playlists,
		"total":     resp.Total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// @Summary Получить плейлист по ID
// @Description Получение информации о плейлисте по его ID
// @Tags playlists
// @Produce json
// @Security BearerAuth
// @Param playlistId path string true "ID плейлиста"
// @Success 200 {object} GetPlaylistResponse "Информация о плейлисте"
// @Failure 401 {object} ErrorResponse "Пользователь не авторизован"
// @Failure 404 {object} ErrorResponse "Плейлист не найден"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/playlists/{playlistId} [get]
func (g *Gateway) getPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playlistId := vars["playlistId"]

	grpcReq := &playlistpb.GetPlaylistRequest{
		PlaylistId: playlistId,
	}

	resp, err := g.playlistClient.GetPlaylist(r.Context(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	if resp.Playlist == nil {
		writeError(w, "Playlist not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Playlist)
}

// @Summary Обновить плейлист
// @Description Обновление информации о плейлисте
// @Tags playlists
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param playlistId path string true "ID плейлиста"
// @Param request body UpdatePlaylistRequest true "Данные для обновления плейлиста"
// @Success 200 {object} map[string]bool "Плейлист успешно обновлен"
// @Failure 400 {object} ErrorResponse "Некорректные данные запроса"
// @Failure 401 {object} ErrorResponse "Пользователь не авторизован"
// @Failure 404 {object} ErrorResponse "Плейлист не найден"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/playlists/{playlistId} [put]
func (g *Gateway) updatePlaylistHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playlistId := vars["playlistId"]

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsPrivate   bool   `json:"is_private"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	grpcReq := &playlistpb.UpdatePlaylistRequest{
		PlaylistId:  playlistId,
		Name:        req.Name,
		Description: req.Description,
		IsPrivate:   req.IsPrivate,
	}

	resp, err := g.playlistClient.UpdatePlaylist(r.Context(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	result := map[string]bool{
		"success": resp.Success,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// @Summary Удалить плейлист
// @Description Удаление плейлиста
// @Tags playlists
// @Produce json
// @Security BearerAuth
// @Param playlistId path string true "ID плейлиста"
// @Success 200 {object} map[string]bool "Плейлист успешно удален"
// @Failure 401 {object} ErrorResponse "Пользователь не авторизован"
// @Failure 404 {object} ErrorResponse "Плейлист не найден"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/playlists/{playlistId} [delete]
func (g *Gateway) deletePlaylistHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playlistId := vars["playlistId"]
	userId := r.Context().Value("user_id").(string)

	grpcReq := &playlistpb.DeletePlaylistRequest{
		PlaylistId: playlistId,
		UserId:     userId,
	}

	resp, err := g.playlistClient.DeletePlaylist(r.Context(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	result := map[string]bool{
		"success": resp.Success,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// @Summary Получить треки плейлиста
// @Description Получение списка треков в плейлисте
// @Tags playlists
// @Produce json
// @Security BearerAuth
// @Param playlistId path string true "ID плейлиста"
// @Param limit query int false "Количество записей на странице" default(50)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {object} GetPlaylistTracksResponse "Список треков плейлиста"
// @Failure 401 {object} ErrorResponse "Пользователь не авторизован"
// @Failure 404 {object} ErrorResponse "Плейлист не найден"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/playlists/{playlistId}/tracks [get]
func (g *Gateway) getPlaylistTracksHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playlistId := vars["playlistId"]

	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	grpcReq := &playlistpb.GetPlaylistTracksRequest{
		PlaylistId: playlistId,
		Limit:      int32(limit),
		Offset:     int32(offset),
	}

	resp, err := g.playlistClient.GetPlaylistTracks(r.Context(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	result := map[string]interface{}{
		"tracks": resp.Tracks,
		"total":  resp.Total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// @Summary Добавить трек в плейлист
// @Description Добавление трека в плейлист
// @Tags playlists
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param playlistId path string true "ID плейлиста"
// @Param request body AddTrackToPlaylistRequest true "Данные для добавления трека"
// @Success 200 {object} map[string]bool "Трек успешно добавлен"
// @Failure 400 {object} ErrorResponse "Некорректные данные запроса"
// @Failure 401 {object} ErrorResponse "Пользователь не авторизован"
// @Failure 404 {object} ErrorResponse "Плейлист или трек не найден"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/playlists/{playlistId}/tracks [post]
func (g *Gateway) addTrackToPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playlistId := vars["playlistId"]
	userId := r.Context().Value("user_id").(string)

	var req struct {
		TrackId string `json:"track_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	grpcReq := &playlistpb.AddTrackToPlaylistRequest{
		PlaylistId: playlistId,
		TrackId:    req.TrackId,
		UserId:     userId,
	}

	resp, err := g.playlistClient.AddTrackToPlaylist(r.Context(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	result := map[string]bool{
		"success": resp.Success,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// @Summary Удалить трек из плейлиста
// @Description Удаление трека из плейлиста
// @Tags playlists
// @Produce json
// @Security BearerAuth
// @Param playlistId path string true "ID плейлиста"
// @Param trackId path string true "ID трека"
// @Success 200 {object} map[string]bool "Трек успешно удален"
// @Failure 401 {object} ErrorResponse "Пользователь не авторизован"
// @Failure 404 {object} ErrorResponse "Плейлист или трек не найден"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/playlists/{playlistId}/tracks/{trackId} [delete]
func (g *Gateway) removeTrackFromPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playlistId := vars["playlistId"]
	trackId := vars["trackId"]
	userId := r.Context().Value("user_id").(string)

	grpcReq := &playlistpb.RemoveTrackFromPlaylistRequest{
		PlaylistId: playlistId,
		TrackId:    trackId,
		UserId:     userId,
	}

	resp, err := g.playlistClient.RemoveTrackFromPlaylist(r.Context(), grpcReq)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	result := map[string]bool{
		"success": resp.Success,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// uploadTrackHandler godoc
//
//	@Summary		Загрузить трек
//	@Description	Загрузка аудиофайла трека через multipart/form-data
//	@Tags			Upload
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file		formData	file	true	"Аудиофайл трека"
//	@Param			artist_ids	formData	[]string	true	"Массив ID артистов"
//	@Param			track_name	formData	string	true	"Название трека"
//	@Param			genre		formData	string	true	"Жанр трека"
//	@Success		200			{object}	object{success=bool,message=string,track_id=string}
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Router			/api/v1/upload/track [post]
func (g *Gateway) uploadTrackHandler(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 100MB)
	err := r.ParseMultipartForm(100 << 20) // 100MB
	if err != nil {
		writeError(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get file
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, "File is required: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	_ = header // Suppress unused variable warning

	// Get metadata
	trackName := r.FormValue("track_name")
	if trackName == "" {
		writeError(w, "track_name is required", http.StatusBadRequest)
		return
	}

	genre := r.FormValue("genre")
	if genre == "" {
		writeError(w, "genre is required", http.StatusBadRequest)
		return
	}

	artistIDsStr := r.Form["artist_ids"]
	if len(artistIDsStr) == 0 {
		writeError(w, "at least one artist_id is required", http.StatusBadRequest)
		return
	}

	// Create gRPC stream
	ctx := r.Context()
	stream, err := g.uploadClient.UploadTrack(ctx)
	if err != nil {
		writeError(w, "Failed to create upload stream: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send metadata first
	metadata := &uploadpb.TrackMetadata{
		ArtistIds: artistIDsStr,
		TrackName: trackName,
		Genre:     genre,
	}

	metadataReq := &uploadpb.UploadTrackRequest{
		Data: &uploadpb.UploadTrackRequest_Metadata{
			Metadata: metadata,
		},
	}

	if err := stream.Send(metadataReq); err != nil {
		writeError(w, "Failed to send metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send file chunks (64KB chunks)
	chunkSize := 64 * 1024 // 64KB
	buffer := make([]byte, chunkSize)

	for {
		n, err := file.Read(buffer)
		if n > 0 {
			chunkReq := &uploadpb.UploadTrackRequest{
				Data: &uploadpb.UploadTrackRequest_Chunk{
					Chunk: buffer[:n],
				},
			}

			if err := stream.Send(chunkReq); err != nil {
				writeError(w, "Failed to send chunk: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			writeError(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Close and receive response
	resp, err := stream.CloseAndRecv()
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	result := map[string]interface{}{
		"success":  resp.Success,
		"message":  resp.Message,
		"track_id": resp.TrackId,
	}

	w.Header().Set("Content-Type", "application/json")
	if resp.Success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(result)
}

// getTracksHandler godoc
//
//	@Summary		Получить список треков
//	@Description	Возвращает список треков с пагинацией
//	@Tags			Tracks
//	@Accept			json
//	@Produce		json
//	@Param			limit		query		int		false	"Количество записей на странице"	default(20)
//	@Param			offset		query		int		false	"Смещение"	default(0)
//	@Param			artist_id	query		string	false	"ID артиста для фильтрации"
//	@Success		200			{object}	object{tracks=[]object,limit=int,offset=int}
//	@Failure		500			{object}	ErrorResponse
//	@Router			/api/v1/tracks [get]
func (g *Gateway) getTracksHandler(w http.ResponseWriter, r *http.Request) {
	// Проксируем запрос к tracks-service HTTP API
	tracksServiceURL := getEnv("TRACKS_SERVICE_URL", "http://tracks-service:8080")
	targetURL, err := url.Parse(tracksServiceURL)
	if err != nil {
		writeError(w, "Invalid tracks service URL", http.StatusInternalServerError)
		return
	}

	// Создаем прокси
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	
	// Модифицируем запрос
	r.URL.Path = "/api/tracks"
	r.URL.Host = targetURL.Host
	r.URL.Scheme = targetURL.Scheme
	r.Host = targetURL.Host

	// Проксируем запрос
	proxy.ServeHTTP(w, r)
}

// searchTracksHandler godoc
//
//	@Summary		Поиск треков
//	@Description	Поиск треков по названию
//	@Tags			Tracks
//	@Accept			json
//	@Produce		json
//	@Param			q		query		string	true	"Поисковый запрос"
//	@Param			limit	query		int		false	"Количество записей на странице"	default(20)
//	@Param			offset	query		int		false	"Смещение"						default(0)
//	@Success		200		{object}	object{query=string,items=[]object,limit=int,offset=int}
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/v1/tracks/search [get]
func (g *Gateway) searchTracksHandler(w http.ResponseWriter, r *http.Request) {
	// Проксируем запрос к tracks-service HTTP API
	tracksServiceURL := getEnv("TRACKS_SERVICE_URL", "http://tracks-service:8080")
	targetURL, err := url.Parse(tracksServiceURL)
	if err != nil {
		writeError(w, "Invalid tracks service URL", http.StatusInternalServerError)
		return
	}

	// Создаем прокси
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	
	// Модифицируем запрос
	r.URL.Path = "/api/tracks/search"
	r.URL.Host = targetURL.Host
	r.URL.Scheme = targetURL.Scheme
	r.Host = targetURL.Host

	// Проксируем запрос
	proxy.ServeHTTP(w, r)
}

// getTrackByIdHandler godoc
//
//	@Summary		Получить трек по ID
//	@Description	Возвращает информацию о треке по его ID
//	@Tags			Tracks
//	@Accept			json
//	@Produce		json
//	@Param			trackId	path		string	true	"ID трека"
//	@Success		200		{object}	object
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/v1/tracks/{trackId} [get]
func (g *Gateway) getTrackByIdHandler(w http.ResponseWriter, r *http.Request) {
	// Проксируем запрос к tracks-service HTTP API
	tracksServiceURL := getEnv("TRACKS_SERVICE_URL", "http://tracks-service:8080")
	targetURL, err := url.Parse(tracksServiceURL)
	if err != nil {
		writeError(w, "Invalid tracks service URL", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	trackId := vars["trackId"]

	// Создаем прокси
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	
	// Модифицируем запрос
	r.URL.Path = "/api/tracks/" + trackId
	r.URL.Host = targetURL.Host
	r.URL.Scheme = targetURL.Scheme
	r.Host = targetURL.Host

	// Проксируем запрос
	proxy.ServeHTTP(w, r)
}
