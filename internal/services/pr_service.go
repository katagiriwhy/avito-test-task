package services

import (
	"context"

	"github.com/katagiriwhy/avito-test-task/internal/models"
	"github.com/katagiriwhy/avito-test-task/internal/storage"
)

type PullRequestService struct {
	db *storage.PostgresStorage
}

func NewPullRequestService(db *storage.PostgresStorage) *PullRequestService {
	return &PullRequestService{db: db}
}

func (p *PullRequestService) CreatePullRequest(ctx context.Context, pr models.PullRequest) error {
	return p.db.CreatePullRequest(ctx, pr)
}

func (p *PullRequestService) MergePullRequest(ctx context.Context, prID string) (*models.PullRequest, error) {
	return p.db.MergePullRequest(ctx, prID)
}

func (p *PullRequestService) ReassignReviewer(ctx context.Context, prID string, reviewerID string) (*models.PullRequest, string, error) {
	return p.db.ReassignReviewer(ctx, prID, reviewerID)
}
