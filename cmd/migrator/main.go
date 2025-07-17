package main

import (
	"context"
	"log/slog"

	di2 "itmostar/internal/app/di"
	"itmostar/internal/config"

	"github.com/defany/db/pkg/postgres"
	"github.com/defany/slogger/pkg/logger/sl"
)

func main() {
	config.MustSetup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	di := di2.DI{}

	db := di.Postgres(ctx)

	log := di.Log(ctx)

	migrator, err := postgres.NewMigrator(db.Pool(), config.PostgresMigrationsPath())
	if err != nil {
		log.Error("failed to setup migrator", sl.ErrAttr(err))

		return
	}

	log.Info("applying migrations")

	upped, err := migrator.Up(ctx)
	if err != nil {
		log.Error("failed to up migrations", sl.ErrAttr(err))

		return
	}

	if len(upped) == 0 {
		log.Info("no new migrations to apply")

		return
	}

	for _, migration := range upped {
		log.Info("migration applied!", slog.String("name", migration.Source.Path))
	}
}
