package tguser

import (
	slerr "github.com/defany/slogger/pkg/err"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

func (h *Handler) alarm(ctx *telegohandler.Context, message telego.Message) error {
	builder := telegoutil.Message(message.Chat.ChatID(), "Ошибка при попытке получения информации про ракетную опасность")

	alarmText, err := h.service.FetchLastAlarmInfo(ctx)
	if err != nil {
		if err := h.sendMessage(ctx, builder); err != nil {
			return slerr.WithSource(err)
		}

		return slerr.WithSource(err)
	}

	builder.WithText(alarmText)
	builder.WithParseMode(telego.ModeHTML)

	if err := h.sendMessage(ctx, builder); err != nil {
		return slerr.WithSource(err)
	}

	return nil
}
