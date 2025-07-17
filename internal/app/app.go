package app

import (
	"context"

	"itmostar/internal/app/di"

	"github.com/defany/platcom/pkg/closer"
	"golang.org/x/sync/errgroup"
)

type App struct {
	di *di.DI
}

func New() *App {
	return &App{}
}

func (a *App) Run(ctx context.Context) error {
	wg, ctx := errgroup.WithContext(ctx)

	closer.SetLogger(a.di.Log(ctx))

	wg.Go(func() error {
		return a.runTelegram(ctx)
	})

	wg.Go(func() error {
		return a.runRiver(ctx)
	})

	return wg.Wait()
}
