package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/MucisSocial/user-service/internal/domain"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

type authService struct {
	userRepo         domain.UserRepository
	refreshTokenRepo domain.RefreshTokenRepository
	jwtService       domain.JWTService
}

func NewAuthService(
	userRepo domain.UserRepository,
	refreshTokenRepo domain.RefreshTokenRepository,
	jwtService domain.JWTService,
) domain.AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
	}
}

func (s *authService) SignUp(ctx context.Context, req *domain.SignUpRequest) (*domain.AuthResponse, error) {
	// Check if email already exists
	emailExists, err := s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if emailExists {
		return nil, ErrEmailExists
	}

	// Check if username already exists
	usernameExists, err := s.userRepo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if usernameExists {
		return nil, ErrUsernameExists
	}

	// Create new user
	user, err := domain.NewUser(req.Username, req.Email, req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Save user to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// Generate tokens
	tokens, err := s.jwtService.GenerateTokens(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Save refresh token
	if err := s.saveRefreshToken(ctx, user.ID, tokens.RefreshToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &domain.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         user,
	}, nil
}

func (s *authService) SignIn(ctx context.Context, req *domain.SignInRequest) (*domain.AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Validate password
	if err := user.ValidatePassword(req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	tokens, err := s.jwtService.GenerateTokens(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Save refresh token
	if err := s.saveRefreshToken(ctx, user.ID, tokens.RefreshToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &domain.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         user,
	}, nil
}

func (s *authService) RefreshTokens(ctx context.Context, refreshToken string) (*domain.AuthTokens, error) {
	// Validate refresh token
	_, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Check if refresh token exists in database
	storedToken, err := s.refreshTokenRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Delete old refresh token
	if err := s.refreshTokenRepo.DeleteByToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to delete old refresh token: %w", err)
	}

	// Generate new tokens
	tokens, err := s.jwtService.GenerateTokens(storedToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Save new refresh token
	if err := s.saveRefreshToken(ctx, storedToken.UserID, tokens.RefreshToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return tokens, nil
}

func (s *authService) ValidateToken(ctx context.Context, token string) (*domain.TokenClaims, error) {
	return s.jwtService.ValidateAccessToken(token)
}

func (s *authService) saveRefreshToken(ctx context.Context, userID, token string) error {
	claims, err := s.jwtService.ValidateRefreshToken(token)
	if err != nil {
		return err
	}

	refreshToken := domain.NewRefreshToken(userID, token, claims.ExpiresAt)
	return s.refreshTokenRepo.Create(ctx, refreshToken)
}
