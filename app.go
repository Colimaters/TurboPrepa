package main

import (
	"context"

	application "TurboPrepa/internal/application"
)

// App is the Wails binding facade. Business operations live in internal/application.
type App struct {
	*application.App
}

func NewApp() *App {
	return &App{App: application.NewApp()}
}

func (a *App) startup(ctx context.Context) {
	a.App.Startup(ctx)
}
