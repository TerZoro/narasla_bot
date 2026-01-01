package scheduler

import (
	"context"
	"narasla_bot/storage"
)

type SchedulerStorage interface {
	ListEnabledUsers(ctx context.Context) ([]storage.User, error)
	PickRandom(ctx context.Context, ownerID int64) (*storage.Page, error)
	Remove(ctx context.Context, p *storage.Page) error
	UpdateLastSendAt(ctx context.Context, ownerID, newTime int64) error
}

type Sender interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}
