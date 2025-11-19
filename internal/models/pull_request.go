package models

import "time"

type PullRequest struct {
	PullRequestID     int        `db:"pull_request_id" json:"pull_request_id"`
	PullRequestName   string     `db:"pull_request_name" json:"pull_request_name"`
	AuthorID          string     `db:"author_id" json:"author_id"`
	Status            prStatus   `db:"status" json:"status"`
	AssignedReviewers []string   `db:"assigned_reviewers" json:"assigned_reviewers"`
	CreatedAt         *time.Time `db:"created_at" json:"created_at,omitempty"`
	MergedAt          *time.Time `db:"merged_at" json:"merged_at,omitempty"`
}

type PullRequestShort struct {
	PullRequestID   string   `db:"pull_request_id" json:"pull_request_id"`
	PullRequestName string   `db:"pull_request_name" json:"pull_request_name"`
	AuthorID        string   `db:"author_id" json:"author_id"`
	Status          prStatus `db:"status" json:"status"`
}

type prStatus string

const (
	Open   prStatus = "OPEN"
	Merged prStatus = "MERGED"
)

func IsValidStatus(status prStatus) bool {
	switch status {
	case Open, Merged:
		return true
	default:
		return false
	}
}
