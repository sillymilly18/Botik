package gpt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"itmostar/internal/config"
	channelscrapper "itmostar/pkg/telegram_scrapper/channel"

	slerr "github.com/defany/slogger/pkg/err"
	"github.com/sashabaranov/go-openai"
)

// Пользователь вводит только "HH:MM" локально для Крымского моста (UTC+3).
// День/месяц/год берём из NOW (UTC) -> конвертим в локальное UTC+3 -> подставляем.
// GPT сам оценивает примерное время в пути по названию места выезда.

// -----------------------------------------------------------------------------
// System prompt: жёсткий, короткий, без воды.
// -----------------------------------------------------------------------------
const rideSystemPrompt = `Ассистент расчёта выезда к Крымскому мосту.
Задача: кратко сказать до скольки выехать, показать текущую очередь, предположить пиковую загруженность.
Сам оцени примерное время в пути по месту выезда (грубо, консервативно). Если мало данных — закладывай запас.
Если входные данные указывают на перекрытие / закрытие / остановку движения / эвакуацию / опасность, то в блоке "Рекомендация" пиши коротко: "Движение перекрыто, подождите открытия" (можно добавить "следите за обновлениями").
НЕ советуй время выезда при закрытии.
"Выехать уже поздно" — только если до целевого времени явно не хватает даже минимально разумного времени в пути (учитывая запас ~30 мин) и движение открыто.
ВСЕГДА выводи все три блока ответа (Рекомендация, Сейчас в очереди, Пиковая), даже при закрытии.
Не используй markdown, кавычки, code block, пояснения. Формат вывода задан пользователем. Возвращай только текст.`

// -----------------------------------------------------------------------------
// User prompt template. Подставляем данными через fmt.Sprintf.
// -----------------------------------------------------------------------------
const rideUserPromptTmpl = `
ДАННЫЕ
------
Сырой текст о ситуации на мосту:
%s

Сырой текст о ракетной / БПЛА обстановке:
%s

Место выезда: %s

Целевое время прибытия (локально у моста, UTC+3): %s
Абсолютное целевое локальное время (дата+время): %s

Текущее время (UTC): %s
Текущее локальное время у моста (UTC+3): %s

До целевого времени осталось: %d минут (~%s)

ИНСТРУКЦИЯ
----------
1. Оцени примерное время в пути на авто от места выезда до Крымского моста (грубо; если не знаешь — консервативно; можно упомянуть "примерно").
2. Если движение закрыто/перекрыто/остановлено — в блоке "Рекомендация" скажи, что движение перекрыто и нужно ждать открытия (подождите, следите за обновлениями). НЕ предлагай время выезда.
3. Если движение открыто: сравни оставшееся время с оценкой пути (+~30 мин буфер). Если успевает — дай "Выехать не позднее HH:MM". Если нет — "Выехать уже поздно".
4. Из текста ситуации вытяни по Керчи и Тамани: очередь (свободно / N авто), ожидание (в часах), либо "данных недостаточно".
5. Предположи пиковую загруженность (типично: ночи легче, утро/вечер плотнее) или "Данных недостаточно".
6. НЕ выводи часовые пояса и НЕ выводи даты в ответе. Только времена HH:MM.
7. НЕ выводи время <= текущему локальному. Если расчёт дал прошлое — "Выехать уже поздно" (если открыто) либо рекомендация ждать (если закрыто).

ФОРМАТ ОТВЕТА (СТРОГО)
----------------------
✅ Рекомендация
<одна строка: напр. Выехать не позднее 02:30, чтобы избежать очередей и успеть. / Выехать уже поздно... / Движение перекрыто, подождите открытия.>

🚦 Сейчас в очереди
📍 Со стороны Керчи: <свободно / N авто, ~M ч / данных недостаточно>
📍 Со стороны Тамани: <... аналогично>

🕒 Пиковая загруженность
<диапазон или "Данных недостаточно">

ТРЕБОВАНИЯ ВЫВОДА
-----------------
Только реальный многострочный текст (ASCII переводы строки). Без markdown, бэктиков, кавычек, разметки, код-блоков или пояснений.`

// -----------------------------------------------------------------------------
// Входные параметры бизнес-слоя.
// -----------------------------------------------------------------------------
type FetchRideRecommendationIn struct {
	From      string // место выезда
	EndRideAt string // "HH:MM" локальное время (UTC+3)
}

// -----------------------------------------------------------------------------
// Основная функция: собирает данные, готовит промпт, вызывает модель.
// -----------------------------------------------------------------------------
func (s *Service) FetchRideRecommendation(ctx context.Context, in FetchRideRecommendationIn) (string, error) {
	// 1. Получаем последние посты с каналов.
	trafficRaw, _, err := channelscrapper.FetchLastPost(ctx, config.GeneralInfoChannel())
	if err != nil {
		return "", slerr.WithSource(err)
	}
	alarmRaw, _, err := channelscrapper.FetchLastPost(ctx, config.AlarmsInfoChannel())
	if err != nil {
		return "", slerr.WithSource(err)
	}

	trafficText := sanitizePost(trafficRaw, 2000, "Нет данных по ситуации на мосту.")
	alarmText := sanitizePost(alarmRaw, 2000, "Нет данных по ракетной обстановке.")

	nowUTC := time.Now().UTC()
	nowUTCStr := nowUTC.Format("2006-01-02 15:04:05")

	bridgeLoc := time.FixedZone("CRIMEA", 3*3600)
	nowLocal := nowUTC.In(bridgeLoc)
	nowLocalStr := nowLocal.Format("2006-01-02 15:04:05")

	targetLocal, parsed := combineLocalDateTime(nowLocal, in.EndRideAt)
	targetLocalStr := "не распознано"
	minsUntil := int64(-1)
	humanUntil := "н/д"
	if parsed {
		targetLocalStr = targetLocal.Format("2006-01-02 15:04:05")
		diff := targetLocal.Sub(nowLocal)
		minsUntil = diff.Milliseconds() / 1000 / 60
		humanUntil = humanizeDiff(diff)
	}

	userPrompt := fmt.Sprintf(
		rideUserPromptTmpl,
		trafficText,
		alarmText,
		strings.TrimSpace(in.From),
		strings.TrimSpace(in.EndRideAt),
		targetLocalStr,
		nowUTCStr,
		nowLocalStr,
		minsUntil,
		humanUntil,
	)

	msgs := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: rideSystemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userPrompt,
		},
	}

	req := openai.ChatCompletionRequest{
		Model:       openai.GPT4o,
		Messages:    msgs,
		Temperature: 0.1,
		Store:       false,
	}

	resp, err := s.gpt.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", slerr.WithSource(err)
	}
	if len(resp.Choices) == 0 {
		return "", slerr.WithSource(fmt.Errorf("empty chat completion response"))
	}

	out := strings.TrimSpace(resp.Choices[0].Message.Content)
	if out == "" {
		return "", slerr.WithSource(fmt.Errorf("empty chat completion message content"))
	}

	return out, nil
}

func sanitizePost(s string, maxRunes int, fallback string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	rs := []rune(s)
	if maxRunes > 0 && len(rs) > maxRunes {
		s = string(rs[:maxRunes]) + "…"
	}
	return s
}

// combineLocalDateTime: берём дату из base (локальное UTC+3 nowLocal) + строку "HH:MM".
func combineLocalDateTime(base time.Time, hhmm string) (time.Time, bool) {
	hhmm = strings.TrimSpace(hhmm)
	if hhmm == "" {
		return base, false
	}

	if !strings.Contains(hhmm, ":") {
		hhmm += ":00"
	}

	t, err := time.ParseInLocation("15:04", hhmm, base.Location())
	if err != nil {
		return base, false
	}

	target := time.Date(base.Year(), base.Month(), base.Day(), t.Hour(), t.Minute(), 0, 0, base.Location())

	return target, true
}

func humanizeDiff(d time.Duration) string {
	mins := int64(d / time.Minute)
	if mins == 0 {
		return "0м"
	}
	sign := ""
	if mins < 0 {
		sign = "-"
		mins = -mins
	}
	h := mins / 60
	m := mins % 60
	if h > 0 {
		return fmt.Sprintf("%s%dч%02dм", sign, h, m)
	}
	return fmt.Sprintf("%s%dm", sign, m)
}
