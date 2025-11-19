package models

type Reviewer struct {
	PullRequestID string `db:"pull_request_id" json:"pull_request_id"`
	ReviewerID    string `db:"reviewer_id" json:"reviewer_id"`
}
