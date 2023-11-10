package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/danyaizm/orders-api/storage"
	"github.com/redis/go-redis/v9"
)

type App struct {
	router    http.Handler
	rdb       *redis.Client
	orderRepo storage.OrderRepo
	config    *Config
}

func New(config *Config) *App {
	app := &App{
		rdb: redis.NewClient(&redis.Options{
			Addr: config.RedisAddress,
		}),
		config: config,
	}

	app.loadRoutes()

	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", a.config.ServerPort),
		Handler: a.router,
	}

	if err := a.rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	defer func() {
		if err := a.rdb.Close(); err != nil {
			fmt.Println("failed to close redis: %w", err)
		}
	}()

	fmt.Printf("Starting server at %s\n", server.Addr)

	ch := make(chan error, 1)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return server.Shutdown(timeout)
	}
}
