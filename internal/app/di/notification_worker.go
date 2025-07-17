package di

import (
	"context"

	workernotify "itmostar/internal/worker/notification"
	diut "itmostar/pkg/di"
)

func (d *DI) NotificationWorker(ctx context.Context) *workernotify.NotificationWorker {
	return diut.Once(ctx, func(ctx context.Context) *workernotify.NotificationWorker {
		return workernotify.New(d.GptService(ctx), d.UserRepo(ctx), d.OpenAiApi(ctx), d.Bot(ctx))
	})
}