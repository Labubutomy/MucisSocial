package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/MucisSocial/user-service/internal/config"
	"github.com/MucisSocial/user-service/internal/domain"
)

type jwtService struct {
	accessSecret      []byte
	refreshSecret     []byte
	accessExpiration  time.Duration
	refreshExpiration time.Duration
}

type AccessTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewJWTService(cfg *config.JWTConfig) domain.JWTService {
	return &jwtService{
		accessSecret:      []byte(cfg.AccessSecret),
		refreshSecret:     []byte(cfg.RefreshSecret),
		accessExpiration:  cfg.AccessExpiration,
		refreshExpiration: cfg.RefreshExpiration,
	}
}

func (j *jwtService) GenerateTokens(userID string) (*domain.AuthTokens, error) {
	// Generate access token
	accessToken, err := j.generateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := j.generateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (j *jwtService) ValidateAccessToken(tokenString string) (*domain.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.accessSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		return &domain.TokenClaims{
			UserID:    claims.UserID,
			ExpiresAt: claims.ExpiresAt.Time,
			IssuedAt:  claims.IssuedAt.Time,
		}, nil
	}

	return nil, errors.New("invalid token")
}

func (j *jwtService) ValidateRefreshToken(tokenString string) (*domain.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.refreshSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		return &domain.TokenClaims{
			UserID:    claims.UserID,
			ExpiresAt: claims.ExpiresAt.Time,
			IssuedAt:  claims.IssuedAt.Time,
		}, nil
	}

	return nil, errors.New("invalid refresh token")
}

func (j *jwtService) RefreshTokens(refreshToken string) (*domain.AuthTokens, error) {
	claims, err := j.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return j.GenerateTokens(claims.UserID)
}

func (j *jwtService) generateAccessToken(userID string) (string, error) {
	now := time.Now()
	claims := &AccessTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessExpiration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.accessSecret)
}

func (j *jwtService) generateRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := &RefreshTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshExpiration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.refreshSecret)
}