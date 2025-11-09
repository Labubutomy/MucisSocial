package domain

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                string             `json:"id" db:"id"`
	Username          string             `json:"username" db:"username"`
	Email             string             `json:"email" db:"email"`
	PasswordHash      string             `json:"-" db:"password_hash"`
	AvatarURL         *string            `json:"avatar_url" db:"avatar_url"`
	MusicTasteSummary *MusicTasteSummary `json:"music_taste_summary"`
	CreatedAt         time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" db:"updated_at"`
}

type MusicTasteSummary struct {
	TopGenres  []string `json:"top_genres" db:"top_genres"`
	TopArtists []string `json:"top_artists" db:"top_artists"`
}

type PublicUser struct {
	ID                string             `json:"id"`
	Username          string             `json:"username"`
	AvatarURL         *string            `json:"avatar_url"`
	MusicTasteSummary *MusicTasteSummary `json:"music_taste_summary"`
}

type SearchHistoryItem struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Query     string    `json:"query" db:"query"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type RefreshToken struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// NewUser creates a new user with hashed password
func NewUser(username, email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// ValidatePassword validates user password
func (u *User) ValidatePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

// ToPublicUser converts User to PublicUser (removes sensitive data)
func (u *User) ToPublicUser() *PublicUser {
	return &PublicUser{
		ID:                u.ID,
		Username:          u.Username,
		AvatarURL:         u.AvatarURL,
		MusicTasteSummary: u.MusicTasteSummary,
	}
}

// NewSearchHistoryItem creates a new search history item
func NewSearchHistoryItem(userID, query string) *SearchHistoryItem {
	return &SearchHistoryItem{
		ID:        uuid.New().String(),
		UserID:    userID,
		Query:     query,
		CreatedAt: time.Now(),
	}
}

// NewRefreshToken creates a new refresh token
func NewRefreshToken(userID, token string, expiresAt time.Time) *RefreshToken {
	return &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
}
