package application

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/katagiriwhy/avito-test-task/internal/handlers"
	"github.com/katagiriwhy/avito-test-task/internal/services"
	"github.com/katagiriwhy/avito-test-task/internal/storage"
)

type Application struct {
	router *gin.Engine
	db     storage.Storage
}

func NewApplication() *Application {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	db := storage.NewPostgresStorage(pool)

	teamService := services.NewTeamService(db)
	userService := services.NewUserService(db)
	prService := services.NewPullRequestService(db)

	router := handlers.NewRoutes(teamService, userService, prService)

	return &Application{
		router: router,
		db:     db,
	}
}

func (a *Application) Run() {
	port := os.Getenv("PORT")

	log.Printf("Server starting on port %s", port)

	if err := a.router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (a *Application) Close() {
	if a.db != nil {
		a.db.Close()
	}
}
