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

const notificationPushTextPrompt = `
TRAFFIC_POST = %s
TRAFFIC_POST_CREATED_AT = %s
ALARM_POST_CREATED_AT = %s
ALERT_POST = %s
NOW = %s (UTC ZONE)

Ты — форматтер для телеграм-бота, который каждый час получает два последних текста из каналов:
	•	traffic_post — про трафик/очереди на Крымском мосту;
	•	alert_post — про тревоги (ракета, БПЛА, учения, закрытия и т.п.).

Твоя задача: разобрать оба текста и вернуть ОДНО готовое сообщение для отправки пользователю в Telegram. Никаких разговоров, пояснений, кода или разметки в фенсах. Просто текст сообщения.

Правила вывода
	1.	Заголовок: 🌉 Крымский мост • {время}, где время бери так:
	•	если в traffic_post явно указано (напр. “17:00”), используй его;
	•	иначе если в alert_post есть время — используй его;
	•	иначе используй поле now из входных данных.
	2.	Выведи две строки трафика: для направления со стороны Тамани и со стороны Крыма (Керчи) и возможности движения с каждой из сторон. Если в тексте есть фразы “со стороны Тамани”, “со стороны Керчи” — используй их. Если только общее “затруднений нет” — отрази это в обеих строках.
Пример:
⛔ Со стороны Тамани: движение закрыто
⛔ Со стороны Крыма: движение закрыто
	3.	Парсь числа:
	•	количество транспортных средств (\d+);
	•	время ожидания: мин / час;
	•	пробку в км.
Если нет явного времени, можно грубо оценить: ожидание_мин ≈ round(vehicles / 5) (т.е. ~300 ТС ≈ 60 мин). Это упрощённая модель.
	4.	Определи уровень загруженности и подбери эмодзи:
	•	0/нет: 🟢
	•	до ~200 ТС или <30м или <1 км: 🟡
	•	до ~500 ТС или 30–90м или 1–5 км: 🟠
	•	больше: 🔴
	•	если явное закрытие: ⛔
	5.	Сформируй краткий совет:
	•	Если оба уровня 🟢: “Можно ехать.”
	•	Если максимум 🟡: “Терпимо, закладывай +{доп_мин}” (доп_мин = макс(ожидание,0) но округли до 10; если ожидание нет — не пиши “+”).
	•	Если есть 🟠 или выше: “Лучше позже” и, если в traffic_post есть окно (“после 20:00”, “до 06:00”), вставь его. Иначе “Лучше подождать снижения трафика.”
	•	Если 🔴: “По возможности не выезжай сейчас.” (плюс окно, если есть).
	•	Если ⛔: “Движение закрыто — жди обновлений.”
	6.	Разбери alert_post:
	•	ключевые слова → иконка:
	•	ракета/ракетная тревога → 🚀
	•	БПЛА/дрон → 🚀
	•	учения/тренировка/стрельбы → 🎯
	•	опасность/тревога без уточнения → ⚠️
	•	Определи перечисленные районы (разделители: запятая, тире, «по»).
	•	Время в тексте (формат HH:MM) покажи в скобках.
	•	Если ничего тревожного нет (“всё спокойно”, пусто, None) → ✅ Безопасность: спокойно.
	7.	Если alert_post содержит однотипные сообщения (несколько строк) — выбери самое серьёзное (ракеты > БПЛА > учения > спокойно). Можно объединить регионы.
	8.	Короткий формат строки угрозы:
	•	🚀/🛸 Тревога: Керчь, Севастополь (13:45)
	•	или одиночные: 🚀 Ракетная: Керчь (13:45).
	•	Если только текст иконки нет — начни с ⚠️.
	9.	Не пиши источники, хэштеги, счётчики просмотров, @username и т.п.
	10.	Не используй markdown-кодовые блоки. Новые строки допускаются.
	11.	Максимум 6 строк в сообщении (лучше 4–5). Пустые строки допустимы для читаемости, но не подряд более одной.
	12.	Всегда выводи в порядке: Заголовок → 2 строки трафика → Совет → Угроза.

Итог — одна многострочная строка без лишних символов вокруг.
`

const notificationPushTTL = 5 * time.Minute

type notificationPushCacheEntry struct {
	text               string
	alarmPost          string
	alarmPostCreated   time.Time
	trafficPost        string
	trafficPostCreated time.Time
	generatedAt        time.Time
}

var (
	notificationPushCacheMu sync.RWMutex
	notificationPushCache   *notificationPushCacheEntry
)

func (s *Service) FetchNotificationPushText(ctx context.Context) (string, error) {
	alarmPost, alarmPostCreatedAt, err := channelscrapper.FetchLastPost(ctx, config.AlarmsInfoChannel())
	if err != nil {
		return "", slerr.WithSource(err)
	}

	trafficPost, trafficPostCreatedAt, err := channelscrapper.FetchLastPost(ctx, config.GeneralInfoChannel())
	if err != nil {
		return "", slerr.WithSource(err)
	}

	notificationPushCacheMu.RLock()
	if c := notificationPushCache; c != nil {
		if time.Since(c.generatedAt) < notificationPushTTL {
			out := c.text
			notificationPushCacheMu.RUnlock()
			return out, nil
		}
	}

	notificationPushCacheMu.RUnlock()

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleDeveloper,
				Content: fmt.Sprintf(
					notificationPushTextPrompt,
					trafficPost,
					trafficPostCreatedAt,
					alarmPost,
					alarmPostCreatedAt,
					time.Now().UTC().String(),
				),
			},
		},
		Store: false,
	}

	resp, err := s.gpt.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", slerr.WithSource(err)
	}

	outgoingMessage := resp.Choices[0].Message.Content

	notificationPushCacheMu.Lock()
	notificationPushCache = &notificationPushCacheEntry{
		text:               outgoingMessage,
		alarmPost:          alarmPost,
		alarmPostCreated:   alarmPostCreatedAt,
		trafficPost:        trafficPost,
		trafficPostCreated: trafficPostCreatedAt,
		generatedAt:        time.Now(),
	}
	notificationPushCacheMu.Unlock()

	return outgoingMessage, nil
}
