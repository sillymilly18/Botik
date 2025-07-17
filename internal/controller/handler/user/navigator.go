package tguser

import (
	slerr "github.com/defany/slogger/pkg/err"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

func (h *Handler) navigator(ctx *telegohandler.Context, message telego.Message) error {
	text := `
🗺 Актуальная карта движения через Крымский мост:

🔗 Яндекс Навигатор:  
https://yandex.ru/maps/-/CDb~5WQq

⚠️ Учти, что данные на картах могут отображаться с задержкой.  
Рекомендуем сверять с живыми отзывами в <a href='https://t.me/+kYA_UqPI8lkxYzUy'>чате</a>.
`

	builder := telegoutil.Message(message.Chat.ChatID(), text)
	builder.WithLinkPreviewOptions(&telego.LinkPreviewOptions{
		IsDisabled: true,
	})
	builder.WithParseMode(telego.ModeHTML)

	if _, err := h.bot.SendMessage(ctx, builder); err != nil {
		return slerr.WithSource(err)
	}

	return nil
}
