package main

import (
	"github.com/katagiriwhy/avito-test-task/internal/application"
)

func main() {
	app := application.NewApplication()

	defer app.Close()

	app.Run()
}
