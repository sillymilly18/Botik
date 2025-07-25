package tguser

import (
	slerr "github.com/defany/slogger/pkg/err"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

func (h *Handler) screening(ctx *telegohandler.Context, message telego.Message) error {
	text := `
🛂 Досмотр на Крымском мосту

✅ Форматы досмотра:
<blockquote expandable>
1. Легковые машины, кроссоверы, минивэны
    Ручной досмотр:
    - Выходите с ручной кладью и проходите через рамку‑металлодетектор; сотрудники проверяют сумки на интроскоп‑сканере (≈65×75см).
    - Автомобиль осматривают с помощью зеркал: днище, багажник, капот, бардачок, ниша с запасным колесом .
Обязательна полная проверка всех дверей, багажника, капота, ниш – чтобы не задерживать процесс, уберите из машины лишние вещи .
Под рукой должны быть документы: паспорт, ПТС или свидетельство о регистрации, а также документы на груз (при наличии) .

2. Микроавтобусы, авто с прицепами и крупные внедорожники
    - Проходят через Инспекционно-Досмотровый Комплекс (ИДК) – «просвечивают» автопоезд.
    - Все пассажиры, включая детей и животных, выходят из машины, проходят в зал досмотра.
    - Водитель остаётся с инспектором для проверки документов.
    - Сканирование группы автомобилей обычно занимает до 20 минут; но в часы пик очередь у ИДК может растянуть ожидание до нескольких часов.
</blockquote>


⏱️ Сколько займет
<blockquote expandable>
    Ручной досмотр: 5–10 минут
    ИДК-сканер: до 20–30 минут
    В пиковое время — задержка до 2–3 часов
</blockquote>

🎫 После досмотра выдают талон «Досмотрено» — предъявляется при повторном въезде

📑 Документы
<blockquote expandable>
    - Паспорт
    - Свидетельство о регистрации транспортного средства (СТС)
    - Водительское удостоверение
    - Полис ОСАГО – Формально не требуют, но может быть проверен по базе.
    - Документы на груз (если перевозите что-либо крупное или подозрительное)
</blockquote>

📦 Багаж
<blockquote expandable>
    - Оптимальный размер багажа – 65х75
</blockquote>

📝 Время и советы
<blockquote expandable>
    - Сам процесс ручного досмотра занимает около 6 минут, а полный процесс — до 1 часа.
    - В часы пик (летний сезон, праздники) очередь к ИДК и ручному досмотру может продлиться несколько часов.
    - Упакуйте вещи компактно — в небольшие сумки, чтобы быстро извлекать и показывать.
</blockquote>
`

	builder := telegoutil.Message(message.Chat.ChatID(), text)
	builder.WithParseMode(telego.ModeHTML)

	if _, err := h.bot.SendMessage(ctx, builder); err != nil {
		return slerr.WithSource(err)
	}

	return nil
}
