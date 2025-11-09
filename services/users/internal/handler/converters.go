package handler

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/MucisSocial/user-service/internal/domain"
	pb "github.com/MucisSocial/user-service/proto/users/v1"
)

// Convert domain models to protobuf messages

func convertUserToPB(user *domain.User) *pb.User {
	pbUser := &pb.User{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}

	if user.AvatarURL != nil {
		pbUser.AvatarUrl = *user.AvatarURL
	}

	if user.MusicTasteSummary != nil {
		pbUser.MusicTasteSummary = &pb.MusicTasteSummary{
			TopGenres:  user.MusicTasteSummary.TopGenres,
			TopArtists: user.MusicTasteSummary.TopArtists,
		}
	}

	return pbUser
}

func convertPublicUserToPB(publicUser *domain.PublicUser) *pb.PublicUser {
	pbUser := &pb.PublicUser{
		Id:       publicUser.ID,
		Username: publicUser.Username,
	}

	if publicUser.AvatarURL != nil {
		pbUser.AvatarUrl = *publicUser.AvatarURL
	}

	if publicUser.MusicTasteSummary != nil {
		pbUser.MusicTasteSummary = &pb.MusicTasteSummary{
			TopGenres:  publicUser.MusicTasteSummary.TopGenres,
			TopArtists: publicUser.MusicTasteSummary.TopArtists,
		}
	}

	return pbUser
}

func convertSearchHistoryToPB(items []*domain.SearchHistoryItem) []*pb.SearchHistoryItem {
	pbItems := make([]*pb.SearchHistoryItem, len(items))
	for i, item := range items {
		pbItems[i] = convertSearchHistoryItemToPB(item)
	}
	return pbItems
}

func convertSearchHistoryItemToPB(item *domain.SearchHistoryItem) *pb.SearchHistoryItem {
	return &pb.SearchHistoryItem{
		Id:        item.ID,
		Query:     item.Query,
		CreatedAt: timestamppb.New(item.CreatedAt),
	}
}
