package di

import (
	"context"

	"itmostar/internal/config"
	"itmostar/internal/service/gpt"
	diut "itmostar/pkg/di"

	"github.com/sashabaranov/go-openai"
)

func (d *DI) OpenAiApi(ctx context.Context) *openai.Client {
	return diut.Once(ctx, func(ctx context.Context) *openai.Client {
		client := openai.NewClient(config.OpenAiToken())

		return client
	})
}

func (d *DI) GptService(ctx context.Context) *gpt.Service {
	return diut.Once(ctx, func(ctx context.Context) *gpt.Service {
		return gpt.New(
			d.OpenAiApi(ctx),
			d.UserRepo(ctx),
		)
	})
}
