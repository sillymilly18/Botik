package gpt

import (
	"context"
	"fmt"
	"sync"
	"time"

	"itmostar/internal/config"
	channelscrapper "itmostar/pkg/telegram_scrapper/channel"

	slerr "github.com/defany/slogger/pkg/err"
	"github.com/sashabaranov/go-openai"
)

var (
	alarmInfoCacheTTL = 5 * time.Minute
	alarmCache        struct {
		sync.RWMutex
		text    string
		expires time.Time
	}
)

const alarmPrompt = `
NOW: %s
POST_CREATED_AT: %s
MESSAGE_TO_PARSE: %s

Проанализируй MESSAGE_TO_PARSE и верни одно-единственное сообщение, ГOТОВОЕ для отправки в Telegram
(plain text; parse_mode не важен). Используй РЕАЛЬНЫЕ переводы строки (ASCII 10) ― никаких "\n", \code-block\
или кавычек вокруг результата.

Формат, если тревога ОБЪЯВЛЕНА:
🚨 Ракетная тревога (объявлена)

📍 Районы: {районы через запятую}
⏱️ Время объявления: {POST_CREATED_AT в MSK, HH:MM}

🔔 Источник: официальные каналы + подтверждение от пользователей
‼️ Рекомендуется:
– Оставаться в транспорте с закрытыми окнами
– Не покидать зону стоянки без необходимости
– Следить за обновлениями

🕒 Последнее обновление: {NOW в MSK +03:00, HH:MM}

Формат, если тревога ОТМЕНЕНА или в тексте в ней ничего не указано (выбери в зависимости от ситуации текст либо перед либо после слеша):
🟢 Ракетная тревога отменена/Ракетной тревоги нет 

🕒 Последнее обновление: {NOW в MSK +03:00, HH:MM}

Правила:
1. «Время объявления» бери из текста; если его нет — используй NOW, конвертировав в MSK (+03:00) и оставив HH:MM.
2. Районы/города перечисляй уникально в порядке первого появления.
3. Добавляй дополнительные рекомендации только если они явно присутствуют в сообщении.
4. Итог — одна многострочная строка без лишних символов вокруг.
`

func (s *Service) FetchLastAlarmInfo(ctx context.Context) (string, error) {
	alarmCache.RLock()

	if time.Now().Before(alarmCache.expires) && alarmCache.text != "" {
		t := alarmCache.text
		alarmCache.RUnlock()
		return t, nil
	}

	alarmCache.RUnlock()

	post, postCreatedAt, err := channelscrapper.FetchLastPost(ctx, config.AlarmsInfoChannel())
	if err != nil {
		return "", slerr.WithSource(err)
	}

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleDeveloper,
				Content: fmt.Sprintf(
					alarmPrompt,
					time.Now().UTC().String(),
					postCreatedAt.UTC().String(),
					post,
				),
			},
		},
		Store: false,
	}

	resp, err := s.gpt.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", slerr.WithSource(err)
	}

	alarmText := resp.Choices[0].Message.Content

	alarmCache.Lock()
	alarmCache.text = alarmText
	alarmCache.expires = time.Now().Add(alarmInfoCacheTTL)
	alarmCache.Unlock()

	return alarmText, nil
}
