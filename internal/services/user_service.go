package services

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	appError "github.com/katagiriwhy/avito-test-task/internal/errors"
	"github.com/katagiriwhy/avito-test-task/internal/models"
	"github.com/katagiriwhy/avito-test-task/internal/storage"
)

type UserService struct {
	storage storage.Storage
}

func NewUserService(storage storage.Storage) *UserService {
	return &UserService{storage: storage}
}

func (u *UserService) SetIsActive(ctx context.Context, userId string, isActive bool) (*models.User, error) {
	updated, err := u.storage.SetIsActive(ctx, userId, isActive)
	if err != nil {
		return nil, err
	}
	if !updated {
		return nil, appError.NewNotFound("NOT_FOUND", "user not found")
	}

	user, err := u.storage.GetUser(ctx, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appError.NewNotFound("NOT_FOUND", "user not found")
		}
		return nil, err
	}

	return user, nil
}

func (u *UserService) GetReviews(ctx context.Context, userId string) ([]models.PullRequestShort, error) {
	_, err := u.storage.GetUser(ctx, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appError.NewNotFound("NOT_FOUND", "user not found")
		}
		return nil, err
	}

	return u.storage.GetReview(ctx, userId)
}
