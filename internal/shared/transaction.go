package shared

import "context"

// Transaction is an opaque scope; the driver type belongs only to DB adapters.
type Transaction interface{ TransactionScope() }
type ExecWithTx interface {
	Exec(context.Context, func(Transaction) error) error
}
