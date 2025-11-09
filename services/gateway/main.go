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

	pb "github.com/MucisSocial/api-gateway/proto/users/v1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Gateway struct {
	userClient pb.UserServiceClient
	jwtSecret  []byte
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	// Connect to gRPC service
	conn, err := grpc.Dial("users-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC service: %v", err)
	}
	defer conn.Close()

	gateway := &Gateway{
		userClient: pb.NewUserServiceClient(conn),
		jwtSecret:  []byte(getEnv("JWT_SECRET", "your-super-secret-access-key-change-in-production")),
	}

	r := mux.NewRouter()

	// CORS middleware
	r.Use(corsMiddleware)

	// Health check
	r.HandleFunc("/__health", healthHandler).Methods("GET")

	// Auth endpoints (no JWT required)
	r.HandleFunc("/api/v1/auth/sign-up", gateway.signUpHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/auth/sign-in", gateway.signInHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/auth/refresh", gateway.refreshHandler).Methods("POST", "OPTIONS")

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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

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
