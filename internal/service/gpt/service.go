package gpt

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type UserProvider interface {
	IsNotificationsEnabled(ctx context.Context, userID int64) (bool, error)
	ToggleNotifications(ctx context.Context, userID int64, isEnabled bool) (err error)
}

type Service struct {
	users UserProvider
	gpt   *openai.Client
}

func New(gpt *openai.Client, users UserProvider) *Service {
	return &Service{
		users: users,
		gpt:   gpt,
	}
}
