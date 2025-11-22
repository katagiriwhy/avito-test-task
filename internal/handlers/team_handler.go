package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	appError "github.com/katagiriwhy/avito-test-task/internal/errors"
	"github.com/katagiriwhy/avito-test-task/internal/models"
	"github.com/katagiriwhy/avito-test-task/internal/services"
	"github.com/katagiriwhy/avito-test-task/internal/validation"
)

type TeamHandler struct {
	service *services.TeamService
}

func NewTeamHandler(service *services.TeamService) *TeamHandler {
	return &TeamHandler{service: service}
}

func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var team models.Team
	if err := c.ShouldBindJSON(&team); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "invalid request body: " + err.Error(),
			},
		})
		return
	}

	if validation.IsEmptyString(team.TeamName) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "team_name is required",
			},
		})
		return
	}

	if len(team.Members) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "members array is required and cannot be empty",
			},
		})
		return
	}

	for i, member := range team.Members {
		if validation.IsEmptyString(member.UserID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "members[" + strconv.Itoa(i) + "].user_id is required",
				},
			})
			return
		}
		if validation.IsEmptyString(member.Username) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "members[" + strconv.Itoa(i) + "].username is required",
				},
			})
			return
		}
	}

	result, err := h.service.CreateUpdateTeam(c.Request.Context(), team)
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

	if result.IsUpdate {
		c.JSON(http.StatusOK, gin.H{"team": result.Team})
	} else {
		c.JSON(http.StatusCreated, gin.H{"team": result.Team})
	}
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
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

	team, err := h.service.GetTeam(c.Request.Context(), teamName)
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

	c.JSON(http.StatusOK, team)
}
