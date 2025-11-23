package services

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
	appError "github.com/katagiriwhy/avito-test-task/internal/errors"
	"github.com/katagiriwhy/avito-test-task/internal/models"
	"github.com/katagiriwhy/avito-test-task/internal/storage"
)

type PullRequestService struct {
	storage storage.Storage
	rng     *rand.Rand
}

func NewPullRequestService(storage storage.Storage) *PullRequestService {
	return &PullRequestService{
		storage: storage,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (p *PullRequestService) CreatePullRequest(ctx context.Context, pr models.PullRequest) (*models.PullRequest, error) {
	exists, err := p.storage.PullRequestExists(ctx, pr.PullRequestID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, appError.NewConflict("PR_EXISTS", "PR id already exists")
	}

	author, err := p.storage.GetUser(ctx, pr.AuthorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appError.NewNotFound("NOT_FOUND", "author not found")
		}
		return nil, err
	}

	candidates, err := p.storage.GetActiveUsersByTeam(ctx, author.TeamName, pr.AuthorID)
	if err != nil {
		return nil, err
	}

	maxReviewers := 2
	if len(candidates) < maxReviewers {
		maxReviewers = len(candidates)
	}

	selectedReviewers := make([]string, 0, maxReviewers)
	if maxReviewers > 0 {
		shuffled := make([]string, len(candidates))
		copy(shuffled, candidates)
		p.rng.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		selectedReviewers = shuffled[:maxReviewers]
	}

	pr.Status = models.Open
	pr.AssignedReviewers = selectedReviewers

	if err := p.storage.CreatePullRequest(ctx, pr); err != nil {
		return nil, err
	}

	return p.storage.GetPullRequestWithReviewers(ctx, pr.PullRequestID)
}

func (p *PullRequestService) MergePullRequest(ctx context.Context, prID string) (*models.PullRequest, error) {
	status, err := p.storage.GetPullRequestStatus(ctx, prID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appError.NewNotFound("NOT_FOUND", "pull request not found")
		}
		return nil, err
	}

	if status == models.Merged {
		return p.storage.GetPullRequestWithReviewers(ctx, prID)
	}

	if err := p.storage.UpdatePullRequestStatus(ctx, prID, models.Merged); err != nil {
		return nil, err
	}

	return p.storage.GetPullRequestWithReviewers(ctx, prID)
}

func (p *PullRequestService) ReassignReviewer(ctx context.Context, prID string, oldReviewerID string) (*models.PullRequest, string, error) {
	pr, err := p.storage.GetPullRequestWithReviewers(ctx, prID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", appError.NewNotFound("NOT_FOUND", "pull request not found")
		}
		return nil, "", err
	}

	if pr.Status == models.Merged {
		return nil, "", appError.NewConflict("PR_MERGED", "cannot reassign on merged PR")
	}

	if pr.AuthorID == oldReviewerID {
		return nil, "", appError.NewConflict("BAD_REQUEST", "cannot reassign reviewer who is the author of this PR")
	}

	assigned, err := p.storage.IsReviewerAssigned(ctx, prID, oldReviewerID)
	if err != nil {
		return nil, "", err
	}
	if !assigned {
		return nil, "", appError.NewConflict("NOT_ASSIGNED", "reviewer is not assigned to this PR")
	}

	team, err := p.storage.GetUserTeam(ctx, oldReviewerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", appError.NewNotFound("NOT_FOUND", "reviewer not found")
		}
		return nil, "", err
	}

	candidates, err := p.storage.GetActiveUsersByTeam(ctx, team, oldReviewerID)
	if err != nil {
		return nil, "", err
	}

	filteredCandidates := make([]string, 0)
	for _, candidate := range candidates {
		if candidate != pr.AuthorID {
			filteredCandidates = append(filteredCandidates, candidate)
		}
	}

	if err := p.storage.DeleteReviewer(ctx, prID, oldReviewerID); err != nil {
		return nil, "", err
	}

	var newReviewer string
	if len(filteredCandidates) > 0 {
		newReviewer = filteredCandidates[p.rng.Intn(len(filteredCandidates))]
		if err := p.storage.AddReviewer(ctx, prID, newReviewer); err != nil {
			return nil, "", err
		}
	}

	updatedPR, err := p.storage.GetPullRequestWithReviewers(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	return updatedPR, newReviewer, nil
}

func (p *PullRequestService) GetPullRequest(ctx context.Context, prID string) (*models.PullRequest, error) {
	pr, err := p.storage.GetPullRequestWithReviewers(ctx, prID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appError.NewNotFound("NOT_FOUND", "pull request not found")
		}
		return nil, err
	}
	return pr, nil
}
