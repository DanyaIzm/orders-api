package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/danyaizm/orders-api/storage"
)

type App struct {
	router  http.Handler
	storage storage.Storage
	config  *Config
}

func New(config *Config, storage storage.Storage) *App {
	app := &App{
		config:  config,
		storage: storage,
	}

	app.loadRoutes()

	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", a.config.ServerPort),
		Handler: a.router,
	}

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
