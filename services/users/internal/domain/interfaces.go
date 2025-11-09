package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdateMusicTaste(ctx context.Context, userID string, summary *MusicTasteSummary) error
	Delete(ctx context.Context, id string) error
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
}

type SearchHistoryRepository interface {
	GetUserSearchHistory(ctx context.Context, userID string, limit int) ([]*SearchHistoryItem, error)
	Add(ctx context.Context, item *SearchHistoryItem) error
	ClearUserHistory(ctx context.Context, userID string) error
	DeleteOldEntries(ctx context.Context, userID string, keepLast int) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteByToken(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}

type JWTService interface {
	GenerateTokens(userID string) (*AuthTokens, error)
	ValidateAccessToken(token string) (*TokenClaims, error)
	ValidateRefreshToken(token string) (*TokenClaims, error)
	RefreshTokens(refreshToken string) (*AuthTokens, error)
}

type TokenClaims struct {
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
}

type AuthService interface {
	SignUp(ctx context.Context, req *SignUpRequest) (*AuthResponse, error)
	SignIn(ctx context.Context, req *SignInRequest) (*AuthResponse, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*AuthTokens, error)
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
}

type UserService interface {
	GetUserByID(ctx context.Context, id string) (*PublicUser, error)
	GetMe(ctx context.Context, userID string) (*User, error)
	UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*User, error)
}

type SearchHistoryService interface {
	GetSearchHistory(ctx context.Context, userID string, limit int) ([]*SearchHistoryItem, error)
	AddSearchHistory(ctx context.Context, userID, query string) (*SearchHistoryItem, error)
	ClearSearchHistory(ctx context.Context, userID string) error
}

// Request/Response types
type SignUpRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Username string `json:"username" validate:"required,min=3,max=50"`
}

type SignInRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}

type UpdateProfileRequest struct {
	Username  *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,url"`
}
