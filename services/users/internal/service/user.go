package service

import (
	"context"
	"fmt"

	"github.com/MucisSocial/user-service/internal/domain"
)

type userService struct {
	userRepo domain.UserRepository
}

func NewUserService(
	userRepo domain.UserRepository,
) domain.UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*domain.PublicUser, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user.ToPublicUser(), nil
}

func (s *userService) GetMe(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID string, req *domain.UpdateProfileRequest) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields if provided
	if req.Username != nil {
		// Check if username already exists
		if *req.Username != user.Username {
			exists, err := s.userRepo.UsernameExists(ctx, *req.Username)
			if err != nil {
				return nil, fmt.Errorf("failed to check username existence: %w", err)
			}
			if exists {
				return nil, ErrUsernameExists
			}
		}
		user.Username = *req.Username
	}

	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

type searchHistoryService struct {
	searchHistoryRepo domain.SearchHistoryRepository
}

func NewSearchHistoryService(searchHistoryRepo domain.SearchHistoryRepository) domain.SearchHistoryService {
	return &searchHistoryService{
		searchHistoryRepo: searchHistoryRepo,
	}
}

func (s *searchHistoryService) GetSearchHistory(ctx context.Context, userID string, limit int) ([]*domain.SearchHistoryItem, error) {
	return s.searchHistoryRepo.GetUserSearchHistory(ctx, userID, limit)
}

func (s *searchHistoryService) AddSearchHistory(ctx context.Context, userID, query string) (*domain.SearchHistoryItem, error) {
	item := domain.NewSearchHistoryItem(userID, query)

	if err := s.searchHistoryRepo.Add(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to add search history: %w", err)
	}

	// Limit search history to 50 items per user
	if err := s.searchHistoryRepo.DeleteOldEntries(ctx, userID, 50); err != nil {
		// Log warning but don't fail the request
		// logger.Warn("failed to clean old search history entries", "error", err)
	}

	return item, nil
}

func (s *searchHistoryService) ClearSearchHistory(ctx context.Context, userID string) error {
	return s.searchHistoryRepo.ClearUserHistory(ctx, userID)
}
