package application

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/katagiriwhy/avito-test-task/internal/storage"
	"os"
)

type Application struct {
	router *gin.Engine
	db     *storage.PostgresStorage
}

func NewApplication() *Application {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		panic(err)
	}

	db := storage.NewPostgresStorage(pool)

	return &Application{
		db: db,
	}
}
