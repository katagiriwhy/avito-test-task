package services

import (
	"context"
	"errors"
	"github.com/katagiriwhy/avito-test-task/internal/models"
	"github.com/katagiriwhy/avito-test-task/internal/storage"
	"strings"
)

type UserService struct {
	db *storage.PostgresStorage
}

func NewUserService(db *storage.PostgresStorage) *UserService {
	return &UserService{db: db}
}

func (u *UserService) SetIsActive(ctx context.Context, userId string, isActive bool) error {

	if strings.TrimSpace(userId) == "" {
		//TODO wrap the mistake
		return errors.New("userId is empty")
	}

	return u.db.SetIsActive(ctx, userId, isActive)
}

func (u *UserService) GetReviews(ctx context.Context, userId string) ([]models.PullRequestShort, error) {
	//TODO the same thing
	if strings.TrimSpace(userId) == "" {
		return nil, errors.New("userId is empty")
	}

	return u.db.GetReview(ctx, userId)
}
