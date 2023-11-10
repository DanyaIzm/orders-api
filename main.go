package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/danyaizm/orders-api/application"
	"github.com/danyaizm/orders-api/storage/redisstorage"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config := application.LoadConfig()
	storage, err := redisstorage.NewRedisStorage(ctx)
	if err != nil {
		panic(err)
	}

	app := application.New(config, storage)

	if err := app.Start(ctx); err != nil {
		panic(err)
	}
}
