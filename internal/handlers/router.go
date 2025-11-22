package handlers

import "github.com/gin-gonic/gin"

func NewRoutes() *gin.Engine {
	router := gin.Default()

	users := router.Group("/users")
	{
		users.POST("setIsActive")
		users.GET("getReview")
	}

	team := router.Group("/team")
	{
		team.GET("get")
		team.POST("add")
	}

	prs := router.Group("/pullRequest")
	{
		prs.POST("create")
		prs.POST("merge")
		prs.POST("reassign")
	}

	return router
}
