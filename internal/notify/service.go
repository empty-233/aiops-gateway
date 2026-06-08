package notify

import (
	"context"
	"errors"
	"log/slog"
)

type Service struct {
	notifiers []Notifier
	logger    *slog.Logger
}

func NewService(notifiers []Notifier, logger *slog.Logger) *Service {
	return &Service{
		notifiers: notifiers,
		logger:    logger,
	}
}

func (s *Service) Notify(ctx context.Context, msg Message) error {
	var errs []error
	for _, notifier := range s.notifiers {
		if notifier == nil {
			continue
		}

		if err := notifier.Notify(ctx, msg); err != nil {
			s.logger.Error("发送通知失败", "channel", notifier.Name(), "error", err)
			errs = append(errs, err)
			continue
		}
		s.logger.Info("发送通知", "channel", notifier.Name(), "title", msg.Title)
	}

	return errors.Join(errs...)
}