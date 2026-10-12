package payments

import (
	"bakery/internal/domains/orders"
	"bakery/internal/shared"
	"context"
)

type Repository interface {
	Insert(context.Context, shared.Transaction, string, string, int64, string) error
	OrderID(context.Context, string) (string, error)
	LockStatus(context.Context, shared.Transaction, string) (string, error)
	HasEvidence(context.Context, shared.Transaction, string) (bool, error)
	Slip(context.Context, shared.Transaction, string, string, []byte) (orders.Evidence, error)
	Submit(context.Context, shared.Transaction, string) error
	Approve(context.Context, shared.Transaction, string, string) error
	Reject(context.Context, shared.Transaction, string, string, string) error
	CancelPending(context.Context, shared.Transaction, string) error
	Evidence(context.Context, string) ([]byte, string, error)
}
