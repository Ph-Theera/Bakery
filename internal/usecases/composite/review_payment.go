package composite

import (
	"bakery/internal/domains/orders"
	catalogout "bakery/internal/ports/port_out/catalog"
	inventoryout "bakery/internal/ports/port_out/inventory"
	notificationout "bakery/internal/ports/port_out/notifications"
	ordersout "bakery/internal/ports/port_out/orders"
	paymentsout "bakery/internal/ports/port_out/payments"
	"bakery/internal/shared"
	"context"
	"fmt"
	"strings"
)

type PaymentReviewArgs struct {
	ExecWithTx          shared.ExecWithTx
	OrderRepository     ordersout.Repository
	PaymentRepository   paymentsout.Repository
	InventoryRepository inventoryout.Repository
	CatalogRepository   catalogout.Repository
	Outbox              notificationout.Outbox
}

func reviewPayment(ctx context.Context, args PaymentReviewArgs, payment, status, actor, reason string) (orders.Data, error) {
	reason = strings.TrimSpace(reason)
	if !shared.UUID(payment) || !shared.UUID(actor) {
		return orders.Data{}, shared.ErrInvalid
	}
	id, e := args.PaymentRepository.OrderID(ctx, payment)
	if e != nil {
		return orders.Data{}, e
	}
	var out orders.Data
	e = args.ExecWithTx.Exec(ctx, func(tx shared.Transaction) error {
		from, e := args.OrderRepository.LockStatus(ctx, tx, id)
		if e != nil {
			return e
		}
		pstate, e := args.PaymentRepository.LockStatus(ctx, tx, payment)
		if e != nil {
			return e
		}
		if e = orders.ValidateReview(from, pstate, status, reason); e != nil {
			return e
		}
		if pstate == status {
			out, e = args.OrderRepository.GetTx(ctx, tx, id)
			return e
		}
		target := "pending_payment"
		if status == "approved" {
			items, e := args.InventoryRepository.Active(ctx, tx, id)
			if e != nil {
				return e
			}
			count, e := args.OrderRepository.ItemCount(ctx, tx, id)
			if e != nil {
				return e
			}
			if len(items) == 0 || count != len(items) {
				return shared.ErrConflict
			}
			for _, i := range items {
				if e = args.CatalogRepository.Deduct(ctx, tx, i.ProductID, i.Quantity); e != nil {
					return e
				}
				if e = args.InventoryRepository.RecordSale(ctx, tx, id, i, actor); e != nil {
					return e
				}
			}
			if e = args.InventoryRepository.Commit(ctx, tx, id); e != nil {
				return e
			}
			if e = args.PaymentRepository.Approve(ctx, tx, payment, actor); e != nil {
				return e
			}
			target = "confirmed"
		} else {
			if e = args.PaymentRepository.Reject(ctx, tx, payment, actor, reason); e != nil {
				return e
			}
		}
		if e = args.OrderRepository.ChangeStatus(ctx, tx, id, target, actor); e != nil {
			return e
		}
		if e = args.OrderRepository.History(ctx, tx, id, from, target, actor, reason); e != nil {
			return e
		}
		out, e = args.OrderRepository.GetTx(ctx, tx, id)
		if e != nil {
			return e
		}
		if status == "approved" {
			return args.Outbox.Enqueue(ctx, tx, id, fmt.Sprintf("TongBai: อนุมัติการชำระเงินแล้ว\nออเดอร์ %s\nลูกค้า %s\nยอด %.2f บาท", out.Number, out.Customer, out.Total))
		}
		return nil
	})
	return out, e
}
