package di

import (
	"context"

	"itmostar/internal/config"
	diut "itmostar/pkg/di"
	"itmostar/pkg/telegram_question"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

func (d *DI) Bot(ctx context.Context) *telego.Bot {
	return diut.Once(ctx, func(ctx context.Context) *telego.Bot {
		bot, err := telego.NewBot(config.TelegramBotToken(), telego.WithDefaultDebugLogger())
		if err != nil {
			d.mustExit(err)
		}

		return bot
	})
}

func (d *DI) BotQuestionManager(ctx context.Context) *telegram_question.Manager {
	return diut.Once(ctx, func(ctx context.Context) *telegram_question.Manager {
		return telegram_question.New(d.Log(ctx), d.Bot(ctx), d.BotHandler(ctx))
	})
}

func (d *DI) BotEventsUpdater(ctx context.Context) <-chan telego.Update {
	return diut.Once(ctx, func(ctx context.Context) <-chan telego.Update {
		bot := d.Bot(ctx)

		updates, err := bot.UpdatesViaLongPolling(ctx, nil)
		if err != nil {
			d.mustExit(err)
		}

		return updates
	})
}

func (d *DI) BotHandler(ctx context.Context) *telegohandler.BotHandler {
	return diut.Once(ctx, func(ctx context.Context) *telegohandler.BotHandler {
		bh, err := telegohandler.NewBotHandler(d.Bot(ctx), d.BotEventsUpdater(ctx))
		if err != nil {
			d.mustExit(err)
		}

		return bh
	})
}
