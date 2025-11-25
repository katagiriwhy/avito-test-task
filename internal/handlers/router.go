package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/katagiriwhy/avito-test-task/internal/services"
)

func NewRoutes(
	teamService *services.TeamService,
	userService *services.UserService,
	prService *services.PullRequestService,
) *gin.Engine {
	router := gin.Default()

	router.Use(corsMiddleware())

	teamHandler := NewTeamHandler(teamService)
	userHandler := NewUserHandler(userService)
	prHandler := NewPullRequestHandler(prService)

	team := router.Group("/team")
	{
		team.GET("/get", teamHandler.GetTeam)
		team.POST("/add", teamHandler.CreateTeam)
	}

	users := router.Group("/users")
	{
		users.POST("/setIsActive", userHandler.SetIsActive)
		users.POST("/deactivate", userHandler.DeactivateTeam)
		users.GET("/getReview", userHandler.GetReview)
	}

	prs := router.Group("/pullRequest")
	{
		prs.POST("/create", prHandler.CreatePullRequest)
		prs.POST("/merge", prHandler.MergePullRequest)
		prs.POST("/reassign", prHandler.ReassignReviewer)
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
