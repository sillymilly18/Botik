package tguser

import (
	slerr "github.com/defany/slogger/pkg/err"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

func (h *Handler) chat(ctx *telegohandler.Context, message telego.Message) error {
	text := `
💬 Перейти в чат

Обсудить очередь, досмотр, задать вопрос или просто поделиться опытом можно в основном чате:

👉 <a href='https://t.me/+kYA_UqPI8lkxYzUy'>КРЫМСКИЙ МОСТ – ЧАТ</a>

Пожалуйста, соблюдайте уважительный тон 🙏
`

	builder := telegoutil.Message(message.Chat.ChatID(), text)
	builder.WithParseMode(telego.ModeHTML)
	builder.WithLinkPreviewOptions(&telego.LinkPreviewOptions{
		IsDisabled: true,
	})

	if _, err := h.bot.SendMessage(ctx, builder); err != nil {
		return slerr.WithSource(err)
	}

	return nil
}
