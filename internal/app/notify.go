package app

import (
	"fmt"
	"log/slog"

	"aiops-gateway/internal/config"
	"aiops-gateway/internal/notify"
	"aiops-gateway/internal/notify/channel"
)

func NewNotifyClient(cfg config.NotifyConfig, logger *slog.Logger) (*notify.Service, error) {
	var notifiers []notify.Notifier

	for _, ch := range cfg.Channels {
		if !ch.Enabled {
			continue
		}

		switch ch.Type {
		case "bark":
			notifiers = append(notifiers, channel.NewBark(
				ch.Name,
				ch.ServerURL,
				ch.Key,
			))
		default:
			return nil, fmt.Errorf("不支持的通知渠道类型: %s", ch.Type)
		}
	}

	return notify.NewService(notifiers, logger), nil
}
