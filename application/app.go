package application

import (
	"context"
	"net/http"
)

type App struct {
	router http.Handler
}

func New() *App {
	return &App{
		router: loadRoutes(),
	}
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    "localhost:3000",
		Handler: a.router,
	}

	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
