package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katagiriwhy/avito-test-task/internal/services"
)

type StatsHandler struct {
	service *services.StatsService
}

func NewStatsHandler(service *services.StatsService) *StatsHandler {
	return &StatsHandler{service: service}
}

func (h *StatsHandler) GetReviewAssignmentStats(c *gin.Context) {
	stats, err := h.service.GetReviewAssignmentStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "failed to load statistics",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"review_assignments": stats,
	})
}
