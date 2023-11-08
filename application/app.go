package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
}

func New() *App {
	return &App{
		router: loadRoutes(),
		rdb:    redis.NewClient(&redis.Options{}),
	}
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    "localhost:3000",
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
