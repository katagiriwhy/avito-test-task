package services

import (
	"context"

	"github.com/katagiriwhy/avito-test-task/internal/models"
	"github.com/katagiriwhy/avito-test-task/internal/storage"
)

type UserService struct {
	db *storage.PostgresStorage
}

func NewUserService(db *storage.PostgresStorage) *UserService {
	return &UserService{db: db}
}

func (u *UserService) SetIsActive(ctx context.Context, userId string, isActive bool) error {
	return u.db.SetIsActive(ctx, userId, isActive)
}

func (u *UserService) GetReviews(ctx context.Context, userId string) ([]models.PullRequestShort, error) {
	return u.db.GetReview(ctx, userId)
}
