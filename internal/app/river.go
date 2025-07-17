package app

import (
	"context"
	"log/slog"

	"github.com/defany/platcom/pkg/closer"
	slerr "github.com/defany/slogger/pkg/err"
)

func (a *App) runRiver(ctx context.Context) error {
	log := a.di.Log(ctx).With(slog.String("instance", "river"))

	river := a.di.RiverClient(ctx)

	closer.Add(func() error {
		return river.Stop(ctx)
	})

	log.Info("go river!")
	if err := river.Start(ctx); err != nil {
		return slerr.WithSource(err)
	}

	return nil
}
