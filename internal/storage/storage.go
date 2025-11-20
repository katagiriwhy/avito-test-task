package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/katagiriwhy/avito-test-task/internal/models"
)

type Storage interface {
	Close()
	CreateUpdateTeam(ctx context.Context, t models.Team) error
	GetTeam(ctx context.Context, teamName string) (*models.Team, error)
	SetIsActive(ctx context.Context, userId string, isActive bool) error
	CreatePullRequest(ctx context.Context, pr models.PullRequest) error
}

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(pool *pgxpool.Pool) *PostgresStorage {
	return &PostgresStorage{
		pool: pool,
	}
}

func (s *PostgresStorage) Close() {
	s.pool.Close()
}

func (s *PostgresStorage) CreateUpdateTeam(ctx context.Context, t models.Team) error {
	tx, err := s.pool.Begin(ctx)

	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const query = `INSERT INTO teams(team_name) VALUES($1) ON CONFLICT DO NOTHING`

	if _, err := tx.Exec(ctx, query, t.TeamName); err != nil {
		return err
	}

	const memberQuery = `INSERT INTO users(user_id, username, team_name, is_active) 
								VALUES($1, $2, $3, $4)
             		ON CONFLICT (user_id) DO UPDATE SET 
                 username = EXCLUDED.username,
                 team_name = EXCLUDED.team_name,
                 is_active = EXCLUDED.is_active`

	for _, m := range t.Members {
		if _, err := tx.Exec(ctx, memberQuery, m.UserID, m.Username, t.TeamName, m.IsActive); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *PostgresStorage) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	const query = `SELECT team_name FROM teams WHERE team_name = $1`

	var team models.Team

	if err := tx.QueryRow(ctx, query, teamName).Scan(&team.TeamName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("team not found")
		}
		return nil, err
	}

	const membersQuery = `SELECT user_id, username, team_name, is_active FROM users WHERE team_name = $1`

	rows, err := tx.Query(ctx, membersQuery, team.TeamName)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	members := make([]models.User, 0)

	for rows.Next() {
		var u models.User

		if err := rows.Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}

		members = append(members, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	team.Members = members

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &team, nil
}

func (s *PostgresStorage) SetIsActive(ctx context.Context, userId string, isActive bool) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	const query = `UPDATE users SET is_active = $1 WHERE user_id = $2`

	tag, err := tx.Exec(ctx, query, isActive, userId)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return tx.Commit(ctx)
}

func (s *PostgresStorage) CreatePullRequest(ctx context.Context, pr models.PullRequest) error {
	tx, err := s.pool.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	const query = `INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status) VALUES ($1, $2, $3, $4)`

	_, err = tx.Exec(ctx, query, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status)
	if err != nil {
		return err
	}

	const reviewerQuery = `INSERT INTO reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)`

	for _, reviewerID := range pr.AssignedReviewers {
		_, err := tx.Exec(ctx, reviewerQuery, pr.PullRequestID, reviewerID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
