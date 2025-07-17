package di

import (
	"context"

	tguser "itmostar/internal/controller/handler/user"
	pguser "itmostar/internal/repo/user"
	diut "itmostar/pkg/di"
)

func (d *DI) TelegramUserHandler(ctx context.Context) *tguser.Handler {
	return diut.Once(ctx, func(ctx context.Context) *tguser.Handler {
		return tguser.New(d.Bot(ctx), d.OpenAiApi(ctx), d.BotQuestionManager(ctx), d.GptService(ctx), d.UserRepo(ctx))
	})
}

func (d *DI) UserRepo(ctx context.Context) *pguser.Repo {
	return diut.Once(ctx, func(ctx context.Context) *pguser.Repo {
		return pguser.New(d.Postgres(ctx))
	})
}
