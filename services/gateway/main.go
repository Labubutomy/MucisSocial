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
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	artistpb "github.com/MucisSocial/api-gateway/proto/artists/v1"
	pb "github.com/MucisSocial/api-gateway/proto/users/v1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	_ "github.com/MucisSocial/api-gateway/docs"
)

type Gateway struct {
	userClient   pb.UserServiceClient
	artistClient artistpb.ArtistServiceClient
	jwtSecret    []byte
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

	gateway := &Gateway{
		userClient:   pb.NewUserServiceClient(userConn),
		artistClient: artistpb.NewArtistServiceClient(artistConn),
		jwtSecret:    []byte(getEnv("JWT_SECRET", "your-super-secret-access-key-change-in-production")),
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

	// Protected endpoints (JWT required)
	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Use(gateway.jwtMiddleware)
	protected.HandleFunc("/me", gateway.getMeHandler).Methods("GET", "OPTIONS")
	protected.HandleFunc("/me", gateway.updateMeHandler).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/me/search-history", gateway.getSearchHistoryHandler).Methods("GET", "OPTIONS")
	protected.HandleFunc("/me/search-history", gateway.addSearchHistoryHandler).Methods("POST", "OPTIONS")
	protected.HandleFunc("/me/search-history", gateway.clearSearchHistoryHandler).Methods("DELETE", "OPTIONS")

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

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			writeError(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := bearerToken[1]
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

	result := map[string]interface{}{
		"query": query,
		"items": resp.Artists,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
