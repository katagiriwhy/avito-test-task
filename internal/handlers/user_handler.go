package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	appError "github.com/katagiriwhy/avito-test-task/internal/errors"
	"github.com/katagiriwhy/avito-test-task/internal/services"
	"github.com/katagiriwhy/avito-test-task/internal/validation"
)

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

type SetIsActiveRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive bool   `json:"is_active"`
}

func (h *UserHandler) SetIsActive(c *gin.Context) {
	var req SetIsActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "invalid request body: " + err.Error(),
			},
		})
		return
	}

	if validation.IsEmptyString(req.UserID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "user_id is required",
			},
		})
		return
	}

	user, err := h.service.SetIsActive(c.Request.Context(), req.UserID, req.IsActive)
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

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *UserHandler) GetReview(c *gin.Context) {
	userID := c.Query("user_id")
	if validation.IsEmptyString(userID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "user_id query parameter is required",
			},
		})
		return
	}

	prs, err := h.service.GetReviews(c.Request.Context(), userID)
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
		"user_id":       userID,
		"pull_requests": prs,
	})
}

func (h *UserHandler) DeactivateTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if validation.IsEmptyString(teamName) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "team_name query parameter is required",
			},
		})
		return
	}

	result, err := h.service.DeactivateTeamUsers(c.Request.Context(), teamName)
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

	c.JSON(http.StatusOK, gin.H{"result": result})
}
