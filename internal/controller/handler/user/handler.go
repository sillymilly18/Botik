package tguser

import (
	"context"

	"itmostar/internal/service/gpt"
	"itmostar/pkg/telegram_question"

	slerr "github.com/defany/slogger/pkg/err"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/sashabaranov/go-openai"
)

type Service interface {
	ToggleNotifications(ctx context.Context, userID int64) (activated bool, err error)
	FetchLastAlarmInfo(ctx context.Context) (string, error)
	FetchLastTrafficInfo(ctx context.Context) (string, error)
	FetchRideRecommendation(ctx context.Context, in gpt.FetchRideRecommendationIn) (string, error)
}

type UserProvider interface {
	CreateIfNotExists(ctx context.Context, userID int64) error
}

type Handler struct {
	gpt     *openai.Client
	qm      *telegram_question.Manager
	bot     *telego.Bot
	service Service
	users   UserProvider
}

func New(bot *telego.Bot, gpt *openai.Client, qm *telegram_question.Manager, service Service, users UserProvider) *Handler {
	return &Handler{
		gpt:     gpt,
		qm:      qm,
		bot:     bot,
		service: service,
		users:   users,
	}
}

func (h *Handler) Setup(handler *telegohandler.BotHandler) {
	handler.Use(func(ctx *telegohandler.Context, update telego.Update) error {
		if update.Message != nil && update.Message.From != nil {
			err := h.users.CreateIfNotExists(ctx, update.Message.From.ID)
			if err != nil {
				return slerr.WithSource(err)
			}
		}

		return ctx.Next(update)
	})

	handler.HandleMessage(h.start, telegohandler.CommandEqual("start"))

	handler.HandleMessage(h.traffic, telegohandler.Or(
		telegohandler.TextContains("трафик"),
		telegohandler.TextContains("Трафик"),
	))

	handler.HandleMessage(h.alarm, telegohandler.Or(
		telegohandler.TextContains("Тревога"),
		telegohandler.TextContains("тревога"),
	))

	handler.HandleMessage(h.chat, telegohandler.Or(
		telegohandler.TextContains("чат"),
		telegohandler.TextContains("Чат"),
	))
	handler.HandleMessage(h.screening, telegohandler.Or(
		telegohandler.TextContains("досмотр"),
		telegohandler.TextContains("Досмотр"),
	))
	handler.HandleMessage(h.toggleNotifications, telegohandler.Or(
		telegohandler.TextContains("уведомления"),
		telegohandler.TextContains("Уведомления"),
		telegohandler.TextContains("Стоп"),
		telegohandler.TextContains("стоп"),
	))
	handler.HandleMessage(h.navigator, telegohandler.Or(telegohandler.TextContains("карта"), telegohandler.TextContains("Карта")))
}

func (h *Handler) sendMessage(ctx context.Context, builder *telego.SendMessageParams) error {
	_, err := h.bot.SendMessage(ctx, builder)
	if err != nil {
		return slerr.WithSource(err)
	}

	return nil
}
