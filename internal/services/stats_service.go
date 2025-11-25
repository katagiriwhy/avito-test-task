package services

import (
	"context"

	"github.com/katagiriwhy/avito-test-task/internal/models"
	"github.com/katagiriwhy/avito-test-task/internal/storage"
)

type StatsService struct {
	storage storage.Storage
}

func NewStatsService(storage storage.Storage) *StatsService {
	return &StatsService{storage: storage}
}

func (s *StatsService) GetReviewAssignmentStats(ctx context.Context) (*models.ReviewAssignmentStats, error) {
	userStats, prStats, err := s.storage.GetReviewAssignmentStats(ctx)
	if err != nil {
		return nil, err
	}

	return &models.ReviewAssignmentStats{
		User:        userStats,
		PullRequest: prStats,
	}, nil
}
