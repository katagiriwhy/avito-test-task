package services

import (
	"context"
	"errors"
	"github.com/katagiriwhy/avito-test-task/internal/models"
	"github.com/katagiriwhy/avito-test-task/internal/storage"
	"strings"
)

type PullRequestService struct {
	db *storage.PostgresStorage
}

func NewPullRequestService(db *storage.PostgresStorage) *PullRequestService {
	return &PullRequestService{db: db}
}

func (p *PullRequestService) CreatePullRequest(ctx context.Context, pr models.PullRequest) error {
	if strings.TrimSpace(pr.PullRequestID) == "" {
		return errors.New("pull request id is empty")
	}

	if strings.TrimSpace(pr.PullRequestName) == "" {
		return errors.New("pull request name is empty")
	}

	return p.db.CreatePullRequest(ctx, pr)
}
