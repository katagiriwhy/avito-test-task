package storage

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	appError "github.com/katagiriwhy/avito-test-task/internal/errors"
	"github.com/katagiriwhy/avito-test-task/internal/models"
)

// TODO maybe i should put it into struct
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

type Storage interface {
	Close()
	CreateUpdateTeam(ctx context.Context, t models.Team) error
	GetTeam(ctx context.Context, teamName string) (*models.Team, error)
	SetIsActive(ctx context.Context, userId string, isActive bool) error
	CreatePullRequest(ctx context.Context, pr models.PullRequest) error
	MergePullRequest(ctx context.Context, prID string) (*models.PullRequest, error)
	GetPullRequestWithReviewers(ctx context.Context, prID string) (*models.PullRequest, error)
	ReassignReviewer(ctx context.Context, prId, oldReviewerID string) (*models.PullRequest, string, error)
	GetReview(ctx context.Context, userID string) ([]models.PullRequestShort, error)
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

func (s *PostgresStorage) MergePullRequest(ctx context.Context, prID string) (*models.PullRequest, error) {
	const query = `SELECT status, merged_at FROM pull_requests WHERE pull_request_id = $1`

	var mergedAt *time.Time
	var status models.PrStatus

	if err := s.pool.QueryRow(ctx, query, prID).Scan(&status, &mergedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("pull request not found")
		}
		return nil, err
	}

	if status == models.Merged {
		pr, err := s.GetPullRequestWithReviewers(ctx, prID)
		if err != nil {
			return nil, err
		}
		return pr, nil
	}

	tx, err := s.pool.Begin(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	const updateQuery = `UPDATE pull_requests SET status = 'MERGED', merged_at = NOW() WHERE pull_request_id = $1`

	tag, err := tx.Exec(ctx, updateQuery, prID)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("pull request not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetPullRequestWithReviewers(ctx, prID)
}

func (s *PostgresStorage) GetPullRequestWithReviewers(ctx context.Context, prID string) (*models.PullRequest, error) {
	const query = `SELECT pull_request_id, pull_request_name, author_id, status, merged_at FROM pull_requests WHERE pull_request_id = $1`

	var pr models.PullRequest

	err := s.pool.QueryRow(ctx, query, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.MergedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("pull request not found")
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

func (s *PostgresStorage) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*models.PullRequest, string, error) {
	var status models.PrStatus
	err := s.pool.QueryRow(ctx,
		`SELECT status FROM pull_requests WHERE pull_request_id=$1`,
		prID,
	).Scan(&status)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", appError.NewNotFound("NOT_FOUND", "pull request not found")
	}
	if err != nil {
		return nil, "", err
	}

	if status == models.Merged {
		return nil, "", appError.NewConflict("PR_MERGED", "cannot reassign on merged PR")
	}

	var exists bool

	const query = `SELECT EXISTS (SELECT 1 FROM reviewers WHERE pull_request_id=$1 AND reviewer_id=$2)`

	if err := s.pool.QueryRow(ctx, query, prID, oldReviewerID).Scan(&exists); err != nil {
		return nil, "", err
	}

	if !exists {
		return nil, "", appError.NewConflict("NOT_ASSIGNED", "reviewer is not assigned to this PR")
	}

	var team string
	err = s.pool.QueryRow(ctx, `SELECT team_name FROM users WHERE user_id=$1`, oldReviewerID).Scan(&team)

	if err != nil {
		return nil, "", err
	}

	const userQuery = `SELECT user_id FROM users WHERE team_name=$1 AND is_active=true AND user_id !=$2`

	rows, err := s.pool.Query(ctx, userQuery, team, oldReviewerID)

	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	users := make([]string, 0)

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, "", err
		}
		users = append(users, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	if len(users) == 0 {
		return nil, "", appError.NewConflict("NO_CANDIDATE", "no active replacement candidate in team")
	}

	newReviewer := users[rng.Intn(len(users))]

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, "", err
	}
	defer tx.Rollback(ctx)

	const deleteReviewerQuery = `DELETE FROM reviewers WHERE pull_request_id=$1 AND reviewer_id=$2`

	tag, err := tx.Exec(ctx, deleteReviewerQuery, prID, oldReviewerID)
	if err != nil {
		return nil, "", err
	}

	if tag.RowsAffected() == 0 {
		return nil, "", appError.NewConflict("NOT_ASSIGNED", "reviewer disappeared")
	}

	const insertReviewerQuery = `INSERT INTO reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)`

	_, err = tx.Exec(ctx, insertReviewerQuery, prID, newReviewer)
	if err != nil {
		return nil, "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, "", err
	}

	pr, err := s.GetPullRequestWithReviewers(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	return pr, newReviewer, nil
}

func (s *PostgresStorage) GetReview(ctx context.Context, prID string) ([]models.PullRequestShort, error) {
	const query = `SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
	FROM reviewers r JOIN pull_requests pr ON r.pull_request_id = pr.pull_request_id WHERE r.reviewer_id =$1
	ORDER BY pr.created_at DESC`

	rows, err := s.pool.Query(ctx, query, prID)
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
