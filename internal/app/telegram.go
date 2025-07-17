package app

import (
	"context"
	"log/slog"

	"github.com/defany/platcom/pkg/closer"
	slerr "github.com/defany/slogger/pkg/err"
	"github.com/mymmrac/telego/telegohandler"
)

func (a *App) runTelegram(ctx context.Context) error {
	log := a.di.Log(ctx).With(slog.String("instance", "telegram"))

	handler := a.di.BotHandler(ctx)

	questionManager := a.di.BotQuestionManager(ctx)

	handler.Use(questionManager.Middleware())

	a.setupTelegramHandlers(ctx, handler)

	closer.Add(func() error {
		return handler.StopWithContext(ctx)
	})

	log.Info("go telegram bot!")

	err := handler.Start()
	if err != nil {
		return slerr.WithSource(err)
	}

	return nil
}

func (a *App) setupTelegramHandlers(ctx context.Context, handler *telegohandler.BotHandler) {
	user := a.di.TelegramUserHandler(ctx)
	user.Setup(handler)
}
