package notifications

import (
	"bakery/internal/domains/notifications"
	"bakery/internal/shared"
	"context"
)

type Outbox interface {
	Pending(context.Context) ([]notifications.Message, error)
	Finish(context.Context, string, bool) error
	Enqueue(context.Context, shared.Transaction, string, string) error
}
type Sender interface {
	Send(context.Context, notifications.Message) error
}
