package handlers

import "github.com/katagiriwhy/avito-test-task/internal/services"

type PullRequestHandler struct {
	service *services.PullRequestService
}

func NewPullRequestHandler(service *services.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{service: service}
}
