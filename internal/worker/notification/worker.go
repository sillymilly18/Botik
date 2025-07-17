package workernotify

import (
	"context"

	slerr "github.com/defany/slogger/pkg/err"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/riverqueue/river"
	"github.com/sashabaranov/go-openai"
)

type Service interface {
	FetchNotificationPushText(ctx context.Context) (string, error)
}

type UserProvider interface {
	FetchNotificationReceivers(ctx context.Context) ([]int64, error)
}

type WorkerArgs struct{}

func (WorkerArgs) Kind() string { return "notification" }

type NotificationWorker struct {
	gpt     *openai.Client
	bot     *telego.Bot
	service Service
	users   UserProvider
	river.WorkerDefaults[WorkerArgs]
}

func New(service Service, users UserProvider, gpt *openai.Client, bot *telego.Bot) *NotificationWorker {
	return &NotificationWorker{
		service: service,
		users:   users,
		bot:     bot,
		gpt:     gpt,
	}
}

func (n *NotificationWorker) Work(ctx context.Context, _ *river.Job[WorkerArgs]) error {
	notificationReceivers, err := n.users.FetchNotificationReceivers(ctx)
	if err != nil {
		return slerr.WithSource(err)
	}

	if len(notificationReceivers) == 0 {
		return nil
	}

	notificationText, err := n.service.FetchNotificationPushText(ctx)
	if err != nil {
		return slerr.WithSource(err)
	}

	builder := telegoutil.Message(telegoutil.ID(0), notificationText)

	for _, receiverID := range notificationReceivers {
		builder.WithChatID(telegoutil.ID(receiverID))

		if _, err := n.bot.SendMessage(ctx, builder); err != nil {
			return slerr.WithSource(err)
		}
	}

	return nil
}
