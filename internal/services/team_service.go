package services

import (
	"context"
	"errors"
	"github.com/katagiriwhy/avito-test-task/internal/models"
	"github.com/katagiriwhy/avito-test-task/internal/storage"
	"strings"
)

type TeamService struct {
	db *storage.PostgresStorage
}

func NewTeamService(db *storage.PostgresStorage) *TeamService {
	return &TeamService{db: db}
}

func (t *TeamService) CreateUpdateTeam(ctx context.Context, team models.Team) error {
	if strings.TrimSpace(team.TeamName) == "" {
		return errors.New("team name is empty")
	}

	if len(team.Members) == 0 {
		return errors.New("team members is empty")
	}

	return t.db.CreateUpdateTeam(ctx, team)
}

func (t *TeamService) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	if strings.TrimSpace(teamName) == "" {
		return nil, errors.New("team name is empty")
	}
	return t.db.GetTeam(ctx, teamName)
}
