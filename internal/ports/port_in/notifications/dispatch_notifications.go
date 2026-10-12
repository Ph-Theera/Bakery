package notifications

import (
	"context"
)

type DispatchNotificationsCommand struct{}

type DispatchNotificationsUsecase interface {
	Handle(context.Context, DispatchNotificationsCommand) error
}
