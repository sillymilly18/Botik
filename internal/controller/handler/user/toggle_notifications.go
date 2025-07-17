package tguser

import (
	"github.com/defany/platcom/pkg/cond"
	slerr "github.com/defany/slogger/pkg/err"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

func (h *Handler) toggleNotifications(ctx *telegohandler.Context, message telego.Message) error {
	enableNotificationsText := `
🔔 Уведомления включены

Вы будете получать краткие обновления о ситуации на Крымском мосту раз в час.
Чтобы отключить — напишите /стоп или /уведомления ещё раз
`

	disableNotificationsText := `
🔔 Уведомления выключены

Вы больше не будете получать краткие обновления о ситуации на Крымском мосту раз в час.
Чтобы включить — напишите /уведомления
`

	activated, err := h.service.ToggleNotifications(ctx, message.From.ID)
	if err != nil {
		return slerr.WithSource(err)
	}

	builder := telegoutil.Message(message.Chat.ChatID(), cond.Ternary(activated, enableNotificationsText, disableNotificationsText))

	if _, err := h.bot.SendMessage(ctx, builder); err != nil {
		return slerr.WithSource(err)
	}

	return nil
}
