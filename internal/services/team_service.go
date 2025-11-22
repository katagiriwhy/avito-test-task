package services

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	appError "github.com/katagiriwhy/avito-test-task/internal/errors"
	"github.com/katagiriwhy/avito-test-task/internal/models"
	"github.com/katagiriwhy/avito-test-task/internal/storage"
)

type TeamService struct {
	storage storage.Storage
}

func NewTeamService(storage storage.Storage) *TeamService {
	return &TeamService{storage: storage}
}

type CreateUpdateResult struct {
	Team     *models.Team
	IsUpdate bool
}

func (t *TeamService) CreateUpdateTeam(ctx context.Context, team models.Team) (*CreateUpdateResult, error) {

	exists, err := t.storage.TeamExists(ctx, team.TeamName)
	if err != nil {
		return nil, err
	}

	if exists {
		existingTeam, err := t.storage.GetTeam(ctx, team.TeamName)
		if err != nil {
			return nil, err
		}

		userID := make(map[string]bool)
		for _, member := range existingTeam.Members {
			userID[member.UserID] = true
		}

		for _, member := range team.Members {
			if !userID[member.UserID] {
				return nil, appError.NewBadRequest("TEAM_EXISTS", "team_name already exists")
			}
		}

		if err := t.storage.UpdateTeam(ctx, team); err != nil {
			return nil, err
		}

		updatedTeam, err := t.storage.GetTeam(ctx, team.TeamName)
		if err != nil {
			return nil, err
		}

		return &CreateUpdateResult{Team: updatedTeam, IsUpdate: true}, nil
	} else {
		if err := t.storage.CreateUpdateTeam(ctx, team); err != nil {
			return nil, err
		}

		createdTeam, err := t.storage.GetTeam(ctx, team.TeamName)
		if err != nil {
			return nil, err
		}

		return &CreateUpdateResult{Team: createdTeam, IsUpdate: false}, nil
	}
}

func (t *TeamService) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	team, err := t.storage.GetTeam(ctx, teamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appError.NewNotFound("NOT_FOUND", "team not found")
		}
		return nil, err
	}
	return team, nil
}
