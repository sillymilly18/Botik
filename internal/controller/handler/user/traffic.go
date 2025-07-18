package tguser

import (
	"log"

	"itmostar/internal/service/gpt"
	"itmostar/pkg/telegram_question"

	slerr "github.com/defany/slogger/pkg/err"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

func (h *Handler) traffic(ctx *telegohandler.Context, message telego.Message) error {
	errText := "Произошла ошибка при получении информации о трафике"
	builder := telegoutil.Message(message.Chat.ChatID(), errText)
	builder.WithParseMode(telego.ModeHTML)

	alarmText, err := h.service.FetchLastTrafficInfo(ctx)
	if err != nil {
		if err := h.sendMessage(ctx, builder); err != nil {
			return slerr.WithSource(err)
		}

		return slerr.WithSource(err)
	}

	builder.WithText(alarmText)

	if err := h.sendMessage(ctx, builder); err != nil {
		return slerr.WithSource(err)
	}

	builder.WithText("🤓 Подсказать тебе лучшее время выезда через Крымский мост?")

	yesNoKb := telegoutil.InlineKeyboard(
		telegoutil.InlineKeyboardRow(
			telegoutil.InlineKeyboardButton("Да").WithCallbackData("yes"),
		),
		telegoutil.InlineKeyboardRow(
			telegoutil.InlineKeyboardButton("Нет").WithCallbackData("no"),
		),
	)

	builder.WithReplyMarkup(yesNoKb)

	if err := h.sendMessage(ctx, builder); err != nil {
		return slerr.WithSource(err)
	}

	builder.WithReplyMarkup(nil)

	h.qm.Question(message.Chat.ID, message.From.ID, func(getter telegram_question.Getter) {
		ctx, recMessage := getter.Update()
		if recMessage.CallbackQuery == nil {
			getter.Next()

			return
		}

		userChoice := recMessage.CallbackQuery.Data
		if userChoice != "yes" {
			builder.WithText(`
Без проблем!
Если передумаешь — просто напиши:
В любом случае ты можешь всегда узнать:
— 🚦 Текущий поток
— 🚨 Воздушную тревогу
— 📍 Очереди со стороны моста
— 🗺 Карту и досмотры

Выбирай нужное ниже 👇
`)

			builder.WithReplyMarkup(nil)

			if err := h.sendMessage(ctx, builder); err != nil {
				return
			}

			return
		}

		builder.WithText("1️⃣ Откуда ты выезжаешь? (например: Анапа, Джанкой, Краснодар)")
		if err := h.sendMessage(ctx, builder); err != nil {
			return
		}

		ctx, recMessage = getter.Update()
		if recMessage.Message == nil {
			getter.Next()

			return
		}

		from := recMessage.Message.Text

		builder.WithText("2️⃣ К скольким часам тебе нужно быть на месте? (Напиши в формате ЧЧ:ММ)")
		if err := h.sendMessage(ctx, builder); err != nil {
			return
		}

		ctx, recMessage = getter.Update()
		if recMessage.Message == nil {
			getter.Next()

			return
		}

		endRideAt := recMessage.Message.Text

		builder.WithText("Секунду, собираю информацию...")
		if err := h.sendMessage(ctx, builder); err != nil {
			return
		}

		outgoingMessage, err := h.service.FetchRideRecommendation(ctx, gpt.FetchRideRecommendationIn{
			From:      from,
			EndRideAt: endRideAt,
		})
		if err != nil {
			log.Println(err)

			return
		}

		builder.WithText(outgoingMessage)
		if err := h.sendMessage(ctx, builder); err != nil {
			return
		}
	})

	return nil
}
