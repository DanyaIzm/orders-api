package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/danyaizm/orders-api/application"
)

func main() {
	app := application.New()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := app.Start(ctx); err != nil {
		panic(err)
	}
}
