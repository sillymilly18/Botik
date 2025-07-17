package gpt

import (
	"context"

	slerr "github.com/defany/slogger/pkg/err"
)

func (s *Service) ToggleNotifications(ctx context.Context, userID int64) (activated bool, err error) {
	isNotificationsEnabled, err := s.users.IsNotificationsEnabled(ctx, userID)
	if err != nil {
		return false, slerr.WithSource(err)
	}

	if isNotificationsEnabled {
		if err := s.users.ToggleNotifications(ctx, userID, false); err != nil {
			return false, slerr.WithSource(err)
		}

		return false, nil
	}

	err = s.users.ToggleNotifications(ctx, userID, true)
	if err != nil {
		return false, slerr.WithSource(err)
	}

	return true, nil
}
