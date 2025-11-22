package services

import (
	"context"

	"github.com/katagiriwhy/avito-test-task/internal/models"
	"github.com/katagiriwhy/avito-test-task/internal/storage"
)

type TeamService struct {
	db *storage.PostgresStorage
}

func NewTeamService(db *storage.PostgresStorage) *TeamService {
	return &TeamService{db: db}
}

func (t *TeamService) CreateUpdateTeam(ctx context.Context, team models.Team) error {
	return t.db.CreateUpdateTeam(ctx, team)
}

func (t *TeamService) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	return t.db.GetTeam(ctx, teamName)
}
