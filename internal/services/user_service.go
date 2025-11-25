package services

import (
	"context"
	"errors"
	"math/rand"

	"github.com/jackc/pgx/v5"
	appError "github.com/katagiriwhy/avito-test-task/internal/errors"
	"github.com/katagiriwhy/avito-test-task/internal/models"
	"github.com/katagiriwhy/avito-test-task/internal/storage"
)

type UserService struct {
	storage storage.Storage
	rng     *rand.Rand
}

func NewUserService(storage storage.Storage, rng *rand.Rand) *UserService {
	return &UserService{
		storage: storage,
		rng:     rng,
	}
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

func (u *UserService) DeactivateTeamUsers(ctx context.Context, teamName string) (*models.TeamDeactivationResult, error) {
	exists, err := u.storage.TeamExists(ctx, teamName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, appError.NewNotFound("NOT_FOUND", "team not found")
	}

	deactivated, affectedPRs, err := u.storage.DeactivateTeamUsers(ctx, teamName)
	if err != nil {
		return nil, err
	}

	uniquePRs := make(map[string]struct{})
	for _, prID := range affectedPRs {
		if _, seen := uniquePRs[prID]; seen {
			continue
		}
		if err := u.rebalanceOpenPR(ctx, prID); err != nil {
			return nil, err
		}
		uniquePRs[prID] = struct{}{}
	}

	return &models.TeamDeactivationResult{
		TeamName:            teamName,
		DeactivatedUsers:    len(deactivated),
		UpdatedPullRequests: len(uniquePRs),
	}, nil
}

func (u *UserService) rebalanceOpenPR(ctx context.Context, prID string) error {
	pr, err := u.storage.GetPullRequestWithReviewers(ctx, prID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	if pr.Status == models.Merged {
		return nil
	}

	teamName, err := u.storage.GetUserTeam(ctx, pr.AuthorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	candidates, err := u.storage.GetActiveUsersByTeam(ctx, teamName, pr.AuthorID)
	if err != nil {
		return err
	}

	assigned := make(map[string]struct{})
	for _, reviewer := range pr.AssignedReviewers {
		assigned[reviewer] = struct{}{}
	}

	filtered := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if _, exists := assigned[candidate]; exists {
			continue
		}
		filtered = append(filtered, candidate)
	}

	if len(filtered) > 0 {
		u.rng.Shuffle(len(filtered), func(i, j int) {
			filtered[i], filtered[j] = filtered[j], filtered[i]
		})
	}

	needed := 2 - len(pr.AssignedReviewers)
	for i := 0; i < len(filtered) && needed > 0; i++ {
		if err := u.storage.AddReviewer(ctx, prID, filtered[i]); err != nil {
			return err
		}
		needed--
	}

	return nil
}
