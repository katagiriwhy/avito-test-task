package models

type ReviewAssignmentStats struct {
	User        map[string]int `json:"user"`
	PullRequest map[string]int `json:"pull_request"`
}
