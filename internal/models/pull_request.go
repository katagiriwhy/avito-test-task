package models

import "time"

type PullRequest struct {
	PullRequestID     int        `db:"pull_request_id" json:"pull_request_id"`
	PullRequestName   string     `db:"pull_request_name" json:"pull_request_name"`
	AuthorID          string     `db:"author_id" json:"author_id"`
	Status            PrStatus   `db:"status" json:"status"`
	AssignedReviewers []string   `db:"assigned_reviewers" json:"assigned_reviewers"`
	CreatedAt         *time.Time `db:"created_at" json:"created_at,omitempty"`
	MergedAt          *time.Time `db:"merged_at" json:"merged_at,omitempty"`
}

type PullRequestReassignResponse struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	ReplacedBy        string   `json:"replaced_by"`
}

type PullRequestShort struct {
	PullRequestID   string   `db:"pull_request_id" json:"pull_request_id"`
	PullRequestName string   `db:"pull_request_name" json:"pull_request_name"`
	AuthorID        string   `db:"author_id" json:"author_id"`
	Status          PrStatus `db:"status" json:"status"`
}

type PrStatus string

const (
	Open   PrStatus = "OPEN"
	Merged PrStatus = "MERGED"
)

func IsValidStatus(status PrStatus) bool {
	switch status {
	case Open, Merged:
		return true
	default:
		return false
	}
}
