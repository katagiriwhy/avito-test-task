package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/katagiriwhy/avito-test-task/internal/models"
)

type Storage interface {
	Close()
	CreateUpdateTeam(ctx context.Context, t models.Team) error
	UpdateTeam(ctx context.Context, t models.Team) error
	GetTeam(ctx context.Context, teamName string) (*models.Team, error)
	DeactivateTeamUsers(ctx context.Context, teamName string) ([]string, []string, error)
	SetIsActive(ctx context.Context, userId string, isActive bool) (bool, error)
	CreatePullRequest(ctx context.Context, pr models.PullRequest) error
	GetPullRequestStatus(ctx context.Context, prID string) (models.PrStatus, error)
	UpdatePullRequestStatus(ctx context.Context, prID string, status models.PrStatus) error
	GetPullRequestWithReviewers(ctx context.Context, prID string) (*models.PullRequest, error)
	GetUserTeam(ctx context.Context, userID string) (string, error)
	GetActiveUsersByTeam(ctx context.Context, teamName string, excludeUserID string) ([]string, error)
	IsReviewerAssigned(ctx context.Context, prID, reviewerID string) (bool, error)
	DeleteReviewer(ctx context.Context, prID, reviewerID string) error
	AddReviewer(ctx context.Context, prID, reviewerID string) error
	GetReview(ctx context.Context, userID string) ([]models.PullRequestShort, error)
	GetUser(ctx context.Context, userID string) (*models.User, error)
	TeamExists(ctx context.Context, teamName string) (bool, error)
	PullRequestExists(ctx context.Context, prID string) (bool, error)
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

func (s *PostgresStorage) UpdateTeam(ctx context.Context, t models.Team) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const memberQuery = `UPDATE users SET 
		username = $1,
		team_name = $2,
		is_active = $3
		WHERE user_id = $4 AND team_name = $2`

	for _, m := range t.Members {
		tag, err := tx.Exec(ctx, memberQuery, m.Username, t.TeamName, m.IsActive, m.UserID)
		if err != nil {
			return err
		}

		if tag.RowsAffected() == 0 {
			return errors.New("user " + m.UserID + " is not a member of team " + t.TeamName)
		}
	}

	return tx.Commit(ctx)
}

func (s *PostgresStorage) DeactivateTeamUsers(ctx context.Context, teamName string) ([]string, []string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	const deactivateQuery = `UPDATE users SET is_active = false WHERE team_name = $1 AND is_active = true RETURNING user_id`

	rows, err := tx.Query(ctx, deactivateQuery, teamName)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	deactivated := make([]string, 0)
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, nil, err
		}
		deactivated = append(deactivated, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	affectedPRs := make([]string, 0)
	if len(deactivated) > 0 {
		const deleteReviewersQuery = `DELETE FROM reviewers r USING pull_requests pr
			WHERE r.pull_request_id = pr.pull_request_id AND pr.status = 'OPEN' AND r.reviewer_id = ANY($1)
			RETURNING r.pull_request_id`

		deleteRows, err := tx.Query(ctx, deleteReviewersQuery, deactivated)
		if err != nil {
			return nil, nil, err
		}
		defer deleteRows.Close()

		for deleteRows.Next() {
			var prID string
			if err := deleteRows.Scan(&prID); err != nil {
				return nil, nil, err
			}
			affectedPRs = append(affectedPRs, prID)
		}
		if err := deleteRows.Err(); err != nil {
			return nil, nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	return deactivated, affectedPRs, nil
}

func (s *PostgresStorage) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	const query = `SELECT team_name FROM teams WHERE team_name = $1`
	var team models.Team

	if err := s.pool.QueryRow(ctx, query, teamName).Scan(&team.TeamName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	const membersQuery = `SELECT user_id, username, team_name, is_active FROM users WHERE team_name = $1`
	rows, err := s.pool.Query(ctx, membersQuery, team.TeamName)
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
	return &team, nil
}

func (s *PostgresStorage) TeamExists(ctx context.Context, teamName string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`
	var exists bool
	err := s.pool.QueryRow(ctx, query, teamName).Scan(&exists)
	return exists, err
}

func (s *PostgresStorage) SetIsActive(ctx context.Context, userId string, isActive bool) (bool, error) {
	const query = `UPDATE users SET is_active = $1 WHERE user_id = $2`
	tag, err := s.pool.Exec(ctx, query, isActive, userId)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (s *PostgresStorage) GetUser(ctx context.Context, userID string) (*models.User, error) {
	const query = `SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1`
	var u models.User
	err := s.pool.QueryRow(ctx, query, userID).Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &u, nil
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
		if _, err := tx.Exec(ctx, reviewerQuery, pr.PullRequestID, reviewerID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *PostgresStorage) PullRequestExists(ctx context.Context, prID string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`
	var exists bool
	err := s.pool.QueryRow(ctx, query, prID).Scan(&exists)
	return exists, err
}

func (s *PostgresStorage) GetPullRequestStatus(ctx context.Context, prID string) (models.PrStatus, error) {
	const query = `SELECT status FROM pull_requests WHERE pull_request_id = $1`
	var status models.PrStatus

	err := s.pool.QueryRow(ctx, query, prID).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", pgx.ErrNoRows
		}
		return "", err
	}

	return status, nil
}

func (s *PostgresStorage) UpdatePullRequestStatus(ctx context.Context, prID string, status models.PrStatus) error {
	const updateQuery = `UPDATE pull_requests SET status = $1, merged_at = NOW() WHERE pull_request_id = $2`
	_, err := s.pool.Exec(ctx, updateQuery, status, prID)
	return err
}

func (s *PostgresStorage) GetPullRequestWithReviewers(ctx context.Context, prID string) (*models.PullRequest, error) {
	const query = `SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at FROM pull_requests WHERE pull_request_id = $1`
	var pr models.PullRequest

	err := s.pool.QueryRow(ctx, query, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	const reviewersQuery = `SELECT reviewer_id FROM reviewers WHERE pull_request_id = $1`
	rows, err := s.pool.Query(ctx, reviewersQuery, pr.PullRequestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviewers := make([]string, 0)
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewerID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	pr.AssignedReviewers = reviewers
	return &pr, nil
}

func (s *PostgresStorage) GetUserTeam(ctx context.Context, userID string) (string, error) {
	const query = `SELECT team_name FROM users WHERE user_id = $1`
	var team string
	err := s.pool.QueryRow(ctx, query, userID).Scan(&team)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", pgx.ErrNoRows
		}
		return "", err
	}
	return team, nil
}

func (s *PostgresStorage) GetActiveUsersByTeam(ctx context.Context, teamName string, excludeUserID string) ([]string, error) {
	const query = `SELECT user_id FROM users WHERE team_name = $1 AND is_active = true AND user_id != $2`
	rows, err := s.pool.Query(ctx, query, teamName, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]string, 0)
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		users = append(users, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *PostgresStorage) IsReviewerAssigned(ctx context.Context, prID, reviewerID string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM reviewers WHERE pull_request_id = $1 AND reviewer_id = $2)`
	var exists bool
	err := s.pool.QueryRow(ctx, query, prID, reviewerID).Scan(&exists)
	return exists, err
}

func (s *PostgresStorage) DeleteReviewer(ctx context.Context, prID, reviewerID string) error {
	const query = `DELETE FROM reviewers WHERE pull_request_id = $1 AND reviewer_id = $2`
	_, err := s.pool.Exec(ctx, query, prID, reviewerID)
	return err
}

func (s *PostgresStorage) AddReviewer(ctx context.Context, prID, reviewerID string) error {
	const query = `INSERT INTO reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)`
	_, err := s.pool.Exec(ctx, query, prID, reviewerID)
	return err
}

func (s *PostgresStorage) GetReview(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	const query = `SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
	FROM reviewers r JOIN pull_requests pr ON r.pull_request_id = pr.pull_request_id WHERE r.reviewer_id = $1
	ORDER BY pr.created_at DESC`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []models.PullRequestShort
	for rows.Next() {
		var pr models.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}
