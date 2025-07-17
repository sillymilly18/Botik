package di

import (
	"context"
	"log/slog"

	diut "itmostar/pkg/di"

	"github.com/defany/slogger/pkg/logger/sl"
)

func (d *DI) Log(ctx context.Context) *slog.Logger {
	return diut.Once(ctx, func(ctx context.Context) *slog.Logger {
		return sl.Default()
	})
}
