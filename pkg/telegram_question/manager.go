package telegram_question

import (
	"log/slog"
	"sync"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

type Getter interface {
	Update() (*telegohandler.Context, telego.Update)
	Next()
}

type updateCtxPair struct {
	ctx     *telegohandler.Context
	update  telego.Update
	proceed chan struct{}
}

type question struct {
	updates chan updateCtxPair
	last    updateCtxPair
}

func (q *question) Update() (*telegohandler.Context, telego.Update) {
	pair := <-q.updates
	q.last = pair
	return pair.ctx, pair.update
}

func (q *question) Next() {
	if q.last.proceed != nil {
		q.last.proceed <- struct{}{}
	}
}

type Manager struct {
	log       *slog.Logger
	bot       *telego.Bot
	handler   *telegohandler.BotHandler
	questions sync.Map // map[int64]*sync.Map
}

func New(log *slog.Logger, bot *telego.Bot, handler *telegohandler.BotHandler) *Manager {
	return &Manager{
		log:     log.With(slog.String("instance", "question_manager")),
		bot:     bot,
		handler: handler,
	}
}

func (m *Manager) Question(chatID, userID int64, cb func(getter Getter)) {
	userMapAny, _ := m.questions.LoadOrStore(chatID, &sync.Map{})
	userMap := userMapAny.(*sync.Map)

	if _, exists := userMap.Load(userID); exists {
		return
	}

	ch := make(chan updateCtxPair, 16)
	q := &question{updates: ch}
	userMap.Store(userID, q)

	go func() {
		defer close(ch)
		cb(q)
		userMap.Delete(userID)
	}()
}

func (m *Manager) Delete(chatID, userID int64) {
	if userMapAny, ok := m.questions.Load(chatID); ok {
		userMap := userMapAny.(*sync.Map)
		userMap.Delete(userID)
	}
}

func (m *Manager) Middleware() telegohandler.Handler {
	return func(ctx *telegohandler.Context, update telego.Update) error {
		var chatID, userID int64

		switch {
		case update.Message != nil && update.Message.From != nil:
			chatID = update.Message.Chat.ID
			userID = update.Message.From.ID
		case update.CallbackQuery != nil:
			userID = update.CallbackQuery.From.ID
			if update.CallbackQuery.Message != nil {
				chatID = update.CallbackQuery.Message.GetChat().ID
			} else {
				chatID = update.CallbackQuery.From.ID
			}
		default:
			return ctx.Next(update)
		}

		userMapAny, ok := m.questions.Load(chatID)
		if !ok {
			return ctx.Next(update)
		}

		userMap := userMapAny.(*sync.Map)
		qAny, ok := userMap.Load(userID)
		if !ok {
			return ctx.Next(update)
		}

		q := qAny.(*question)

		pair := updateCtxPair{
			ctx:     ctx.WithoutCancel(),
			update:  update,
			proceed: make(chan struct{}, 1),
		}

		log := m.log.With(slog.Int64("chat_id", chatID), slog.Int64("user_id", userID))

		select {
		case q.updates <- pair:
		default:
			log.Warn("update dropped: channel full")
			return nil
		}

		select {
		case <-pair.proceed:
			log.Info("proceed received")
		}

		return ctx.Next(update)
	}
}
