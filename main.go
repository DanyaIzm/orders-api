package main

import (
	"context"

	"github.com/danyaizm/orders-api/application"
)

func main() {
	app := application.New()

	if err := app.Start(context.TODO()); err != nil {
		panic(err)
	}
}
