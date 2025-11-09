package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/MucisSocial/user-service/internal/domain"
	pb "github.com/MucisSocial/user-service/proto/users/v1"
)

type UserServiceHandler struct {
	pb.UnimplementedUserServiceServer
	authService          domain.AuthService
	userService          domain.UserService
	searchHistoryService domain.SearchHistoryService
}

func NewUserServiceHandler(
	authService domain.AuthService,
	userService domain.UserService,
	searchHistoryService domain.SearchHistoryService,
) *UserServiceHandler {
	return &UserServiceHandler{
		authService:          authService,
		userService:          userService,
		searchHistoryService: searchHistoryService,
	}
}

// Authentication methods
func (h *UserServiceHandler) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {
	authReq := &domain.SignInRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	result, err := h.authService.SignIn(ctx, authReq)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "authentication failed: %v", err)
	}

	return &pb.SignInResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User:         convertUserToPB(result.User),
	}, nil
}

func (h *UserServiceHandler) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	authReq := &domain.SignUpRequest{
		Email:    req.Email,
		Password: req.Password,
		Username: req.Username,
	}

	result, err := h.authService.SignUp(ctx, authReq)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "registration failed: %v", err)
	}

	return &pb.SignUpResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User:         convertUserToPB(result.User),
	}, nil
}

func (h *UserServiceHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	tokens, err := h.authService.RefreshTokens(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "token refresh failed: %v", err)
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *UserServiceHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := h.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:     true,
		UserId:    claims.UserID,
		ExpiresAt: timestamppb.New(claims.ExpiresAt),
	}, nil
}

// User management methods
func (h *UserServiceHandler) GetMe(ctx context.Context, req *pb.GetMeRequest) (*pb.GetMeResponse, error) {
	user, err := h.userService.GetMe(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &pb.GetMeResponse{
		User: convertUserToPB(user),
	}, nil
}

func (h *UserServiceHandler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	updateReq := &domain.UpdateProfileRequest{}

	if req.Username != nil {
		updateReq.Username = req.Username
	}
	if req.AvatarUrl != nil {
		updateReq.AvatarURL = req.AvatarUrl
	}

	user, err := h.userService.UpdateProfile(ctx, req.UserId, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "profile update failed: %v", err)
	}

	return &pb.UpdateProfileResponse{
		User: convertUserToPB(user),
	}, nil
}

func (h *UserServiceHandler) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	publicUser, err := h.userService.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &pb.GetUserByIdResponse{
		User: convertPublicUserToPB(publicUser),
	}, nil
}

// Search history methods
func (h *UserServiceHandler) GetSearchHistory(ctx context.Context, req *pb.GetSearchHistoryRequest) (*pb.GetSearchHistoryResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}

	items, err := h.searchHistoryService.GetSearchHistory(ctx, req.UserId, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get search history: %v", err)
	}

	return &pb.GetSearchHistoryResponse{
		Items: convertSearchHistoryToPB(items),
	}, nil
}

func (h *UserServiceHandler) AddSearchHistory(ctx context.Context, req *pb.AddSearchHistoryRequest) (*pb.AddSearchHistoryResponse, error) {
	item, err := h.searchHistoryService.AddSearchHistory(ctx, req.UserId, req.Query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add search history: %v", err)
	}

	return &pb.AddSearchHistoryResponse{
		Item: convertSearchHistoryItemToPB(item),
	}, nil
}

func (h *UserServiceHandler) ClearSearchHistory(ctx context.Context, req *pb.ClearSearchHistoryRequest) (*pb.ClearSearchHistoryResponse, error) {
	err := h.searchHistoryService.ClearSearchHistory(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to clear search history: %v", err)
	}

	return &pb.ClearSearchHistoryResponse{
		Success: true,
	}, nil
}
