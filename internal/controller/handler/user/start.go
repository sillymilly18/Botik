package tguser

import (
	slerr "github.com/defany/slogger/pkg/err"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

func (h *Handler) start(ctx *th.Context, message telego.Message) error {
	messageText := `
👋 Привет! Я бот, который поможет тебе безопасно и удобно пересечь Крымский мост.

🛣 Я подскажу:
— Текущую загруженность моста 
— Время ожидания и длину очереди  
— Информацию о воздушных тревогах  
— Актуальную карту и погодные условия

📤 Также ты можешь сам сообщить о пробке и помочь другим.

Выбери действие с помощью кнопок ниже ⬇️
`

	builder := telegoutil.Message(message.Chat.ChatID(), messageText)

	kb := telegoutil.Keyboard(
		telegoutil.KeyboardRow(
			telegoutil.KeyboardButton("🚦 Трафик"),
			telegoutil.KeyboardButton("🗺 Карта"),
		),
		telegoutil.KeyboardRow(
			telegoutil.KeyboardButton("🔔 Уведомления"),
		),
		telegoutil.KeyboardRow(
			telegoutil.KeyboardButton("🚨 Тревога"),
			telegoutil.KeyboardButton("🛂 Досмотр"),
		),
	)
	kb.WithResizeKeyboard()

	builder.WithReplyMarkup(kb)

	if _, err := h.bot.SendMessage(ctx, builder); err != nil {
		return slerr.WithSource(err)
	}

	return nil
}
