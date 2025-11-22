package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	appError "github.com/katagiriwhy/avito-test-task/internal/errors"
	"github.com/katagiriwhy/avito-test-task/internal/models"
	"github.com/katagiriwhy/avito-test-task/internal/services"
	"github.com/katagiriwhy/avito-test-task/internal/validation"
)

type PullRequestHandler struct {
	service *services.PullRequestService
}

func NewPullRequestHandler(service *services.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{service: service}
}

type CreatePullRequestRequest struct {
	PullRequestID   string `json:"pull_request_id" binding:"required"`
	PullRequestName string `json:"pull_request_name" binding:"required"`
	AuthorID        string `json:"author_id" binding:"required"`
}

type MergePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
}

type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
	OldUserID     string `json:"old_user_id" binding:"required"`
}

func (h *PullRequestHandler) CreatePullRequest(c *gin.Context) {
	var req CreatePullRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "invalid request body: " + err.Error(),
			},
		})
		return
	}

	if validation.IsEmptyString(req.PullRequestID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "pull_request_id is required",
			},
		})
		return
	}

	if validation.IsEmptyString(req.PullRequestName) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "pull_request_name is required",
			},
		})
		return
	}

	if validation.IsEmptyString(req.AuthorID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "author_id is required",
			},
		})
		return
	}

	pr := models.PullRequest{
		PullRequestID:   req.PullRequestID,
		PullRequestName: req.PullRequestName,
		AuthorID:        req.AuthorID,
	}

	createdPR, err := h.service.CreatePullRequest(c.Request.Context(), pr)
	if err != nil {
		var appErr *appError.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "internal server error",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"pr": createdPR})
}

func (h *PullRequestHandler) MergePullRequest(c *gin.Context) {
	var req MergePullRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "invalid request body: " + err.Error(),
			},
		})
		return
	}

	if validation.IsEmptyString(req.PullRequestID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "pull_request_id is required",
			},
		})
		return
	}

	pr, err := h.service.MergePullRequest(c.Request.Context(), req.PullRequestID)
	if err != nil {
		var appErr *appError.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "internal server error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pr": pr})
}

func (h *PullRequestHandler) ReassignReviewer(c *gin.Context) {
	var req ReassignReviewerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "invalid request body: " + err.Error(),
			},
		})
		return
	}

	if validation.IsEmptyString(req.PullRequestID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "pull_request_id is required",
			},
		})
		return
	}

	if validation.IsEmptyString(req.OldUserID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "old_reviewer_id is required",
			},
		})
		return
	}

	pr, newReviewer, err := h.service.ReassignReviewer(c.Request.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		var appErr *appError.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "internal server error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pr":          pr,
		"replaced_by": newReviewer,
	})
}
