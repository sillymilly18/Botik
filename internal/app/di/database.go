package di

import (
	"context"
	"log/slog"
	"time"

	"itmostar/internal/config"
	workernotify "itmostar/internal/worker/notification"
	diut "itmostar/pkg/di"

	"github.com/defany/db/v2/postgres"
	txman "github.com/defany/db/v2/tx_manager"
	"github.com/defany/platcom/pkg/closer"
	"github.com/defany/slogger/pkg/logger/sl"
	"github.com/gookit/goutil/timex"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

func (d *DI) Postgres(ctx context.Context) postgres.Postgres {
	return diut.Once(ctx, func(ctx context.Context) postgres.Postgres {
		pcfg := postgres.NewConfig(
			config.PostgresUsername(),
			config.PostgresPassword(),
			config.PostgresHost(),
			config.PostgresPort(),
			config.PostgresDatabase(),
		)

		pcfg.WithConnAmount(100)
		pcfg.WithMinConnAmount(5)
		pcfg.WithMaxConnLifetime(time.Hour)
		pcfg.WithMaxConnIdleTime(10 * time.Minute)
		pcfg.WithHealthCheckPeriod(2 * time.Minute)

		pg, err := postgres.NewPostgres(ctx, d.Log(ctx), pcfg)
		if err != nil {
			d.mustExit(err)
		}

		closer.Add(func() error {
			d.Log(ctx).Info("shutting down postgres")

			pg.Close()

			return nil
		})

		return pg
	})
}

func (d *DI) TxManager(ctx context.Context) txman.TxManager {
	return diut.Once(ctx, func(ctx context.Context) txman.TxManager {
		return txman.New(d.Postgres(ctx))
	})
}

func (d *DI) RiverClient(ctx context.Context) *river.Client[pgx.Tx] {
	return diut.Once(ctx, func(ctx context.Context) *river.Client[pgx.Tx] {
		pool := d.Postgres(ctx).Pool()

		workers := river.NewWorkers()
		river.AddWorker[workernotify.WorkerArgs](workers, d.NotificationWorker(ctx))

		periodicJobs := []*river.PeriodicJob{
			river.NewPeriodicJob(
				river.PeriodicInterval(time.Hour),
				func() (river.JobArgs, *river.InsertOpts) {
					return workernotify.WorkerArgs{}, nil
				},
				nil,
			),
		}

		cfg := &river.Config{
			Queues: map[string]river.QueueConfig{
				river.QueueDefault: {
					MaxWorkers: 5,
				},
			},
			Workers:      workers,
			PeriodicJobs: periodicJobs,
			Logger: sl.NewSlogLogger(sl.Slog{
				Level:     slog.LevelWarn,
				AddSource: false,
				Format:    "pretty",
			}),
			CancelledJobRetentionPeriod: timex.Day * 7,
			CompletedJobRetentionPeriod: timex.Day * 2,
			DiscardedJobRetentionPeriod: timex.Day * 7,
		}

		client, err := river.NewClient[pgx.Tx](riverpgxv5.New(pool), cfg)
		if err != nil {
			d.mustExit(err)
		}

		return client
	})
}
