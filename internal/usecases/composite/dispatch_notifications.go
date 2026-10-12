package composite

import (
	portin "bakery/internal/ports/port_in/notifications"
	portout "bakery/internal/ports/port_out/notifications"
	"context"
)

type DispatchNotificationsArgs struct {
	Outbox portout.Outbox
	Sender portout.Sender
}
type dispatchNotificationsHandler struct{ args DispatchNotificationsArgs }

func NewDispatchNotificationsUsecase(args DispatchNotificationsArgs) portin.DispatchNotificationsUsecase {
	return &dispatchNotificationsHandler{args: args}
}
func (h *dispatchNotificationsHandler) Handle(ctx context.Context, _ portin.DispatchNotificationsCommand) error {
	messages, e := h.args.Outbox.Pending(ctx)
	if e != nil {
		return e
	}
	for _, m := range messages {
		err := h.args.Sender.Send(ctx, m)
		if e = h.args.Outbox.Finish(ctx, m.ID, err == nil); e != nil {
			return e
		}
	}
	return nil
}
