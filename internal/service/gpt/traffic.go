package gpt

import (
	"context"
	"fmt"
	"sync"
	"time"

	"itmostar/internal/config"
	"itmostar/pkg/format"
	channelscrapper "itmostar/pkg/telegram_scrapper/channel"

	"github.com/bytedance/sonic"
	slerr "github.com/defany/slogger/pkg/err"
	"github.com/sashabaranov/go-openai"
)

type trafficPromptPlace struct {
	IsBusy bool `json:"is_busy"`
	Closed bool `json:"closed"`
	Cars   *int `json:"cars"`
	Wait   *int `json:"wait"`
}

type trafficPromptResult struct {
	Tamani  trafficPromptPlace `json:"tamani"`
	Kerch   trafficPromptPlace `json:"kerch"`
	Alert   bool               `json:"alert"`
	AddedAt time.Time          `json:"added_at"`
}

const trafficPrompt = `
Парси текст и выдай единственную строку-JSON (никаких тегов, код-блоков и переносов).
Формат:

{
“tamani”: { “is_busy”: true | false, "closed": true | false, "“cars”: N (если is_busy), “wait”: N | null },
“kerch”: { “is_busy”: true | false, "closed": true | false,  “cars”: N (если is_busy), “wait”: N | null },
“alert”: true | false,
“added_at”: “YYYY-MM-DDTHH:MM:SS+03:00”
}

Правила:
• «затруднений нет» → is_busy=false, wait=null.
• «движение закрыто» → is_busy=false, wait=null, closed = true.
• Если указана очередь — is_busy=true; заполни cars и wait (в wait только число часов).
• alert=false, если явно сказано, что тревоги нет; иначе true только при явной тревоге.
• added_at: возьми дату моего обращения %s (UTC), замени в ней только часы и минуты на время из текста, установи часовой пояс +03:00 и верни в RFC 3339 (Go time.Time).

Ответ — ровно одна строка JSON.

Текст:
%s
`

var (
	trafficCacheTTL = 5 * time.Minute
	trafficCache    struct {
		sync.RWMutex
		result  trafficPromptResult
		expires time.Time
	}
)

func (s *Service) FetchLastTrafficInfo(ctx context.Context) (string, error) {
	var result trafficPromptResult
	now := time.Now()

	trafficCache.RLock()
	if now.Before(trafficCache.expires) {
		result = trafficCache.result
		trafficCache.RUnlock()
	} else {
		trafficCache.RUnlock()

		post, _, err := channelscrapper.FetchLastPost(ctx, config.GeneralInfoChannel())
		if err != nil {
			return "", slerr.WithSource(err)
		}

		req := openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleDeveloper,
					Content: fmt.Sprintf(trafficPrompt, time.Now().UTC().String(), post),
				},
			},
			Store: false,
		}

		resp, err := s.gpt.CreateChatCompletion(ctx, req)
		if err != nil {
			return "", slerr.WithSource(err)
		}

		if err := sonic.UnmarshalString(resp.Choices[0].Message.Content, &result); err != nil {
			return "", slerr.WithSource(err)
		}

		trafficCache.Lock()
		trafficCache.result = result
		trafficCache.expires = now.Add(trafficCacheTTL)
		trafficCache.Unlock()
	}

	infoMaker := func(place string, in trafficPromptPlace) string {
		if in.Closed {
			return fmt.Sprintf("📍 Въезд со стороны %s: закрыт", place)
		}

		if !in.IsBusy {
			return fmt.Sprintf("📍 Въезд со стороны %s: свободно", place)
		}

		var approxTime string
		if wait := in.Wait; wait != nil {
			approxTime = fmt.Sprintf(", ~%d %s ожидания", *wait, format.Declension(*wait, "час", "часа", "часов"))
		}

		var carsText string
		if cars := in.Cars; cars != nil {
			carsText = fmt.Sprintf("%d авто", *cars)
		}

		return fmt.Sprintf("📍 Въезд с %s: %s%s", place, carsText, approxTime)
	}

	kerchInfo := infoMaker("Керчи", result.Kerch)
	tamaniInfo := infoMaker("Тамани", result.Tamani)

	alarmText, err := s.FetchLastAlarmInfo(ctx)
	if err != nil {
		return "", slerr.WithSource(err)
	}

	finalText := `
🟠 Крымский мост сейчас:

%s
%s

🚨 Информация о ракетной опасности:
<blockquote expandable>
%s
</blockquote>
`
	finalText = fmt.Sprintf(finalText, kerchInfo, tamaniInfo, alarmText)

	return finalText, nil
}
