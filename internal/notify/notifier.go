package notify

import (
	"context"
)

type Notifier interface {
	Name() string
	Notify(ctx context.Context, msg Message) error
}
