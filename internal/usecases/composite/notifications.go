package composite

import "context"

type Message struct{ ID, Text string }
type Outbox interface {
	Pending(context.Context) ([]Message, error)
	Finish(context.Context, string, bool) error
}
type Sender interface {
	Send(context.Context, Message) error
}
type Notifications struct {
	Repo   Outbox
	Sender Sender
}

func (n Notifications) Dispatch(ctx context.Context) error {
	messages, err := n.Repo.Pending(ctx)
	if err != nil {
		return err
	}
	for _, m := range messages {
		e := n.Sender.Send(ctx, m)
		if err = n.Repo.Finish(ctx, m.ID, e == nil); err != nil {
			return err
		}
	}
	return nil
}
