package composite

import (
	"bakery/internal/domains/orders"
	portin "bakery/internal/ports/port_in/orders"
	inventoryout "bakery/internal/ports/port_out/inventory"
	ordersout "bakery/internal/ports/port_out/orders"
	paymentsout "bakery/internal/ports/port_out/payments"
	usersout "bakery/internal/ports/port_out/users"
	"bakery/internal/services/checkout"
	"bakery/internal/shared"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
)

type CreateOrderArgs struct {
	ExecWithTx          shared.ExecWithTx
	OrderRepository     ordersout.Repository
	UserRepository      usersout.Repository
	PaymentRepository   paymentsout.Repository
	InventoryRepository inventoryout.Repository
	Pricing             checkout.Pricing
}
type createOrderHandler struct{ args CreateOrderArgs }

func NewCreateOrderUsecase(args CreateOrderArgs) portin.CreateOrderUsecase {
	return &createOrderHandler{args: args}
}
func (h *createOrderHandler) Handle(ctx context.Context, c portin.CreateOrderCommand) (orders.Data, error) {
	if e := orders.ValidateCreate(c); e != nil {
		return orders.Data{}, e
	}
	raw, e := json.Marshal(struct {
		Command orders.Command
		UserID  string
	}{c, c.UserID})
	if e != nil {
		return orders.Data{}, e
	}
	sum := sha256.Sum256(raw)
	hash := hex.EncodeToString(sum[:])
	var out orders.Data
	e = h.args.ExecWithTx.Exec(ctx, func(tx shared.Transaction) error {
		if e := h.args.OrderRepository.LockKey(ctx, tx, c.Key); e != nil {
			return e
		}
		existing, oldHash, e := h.args.OrderRepository.FindByKey(ctx, tx, c.Key)
		if e == nil {
			if oldHash != hash {
				return shared.ErrConflict
			}
			out = existing
			return nil
		}
		if !errors.Is(e, shared.ErrNotFound) {
			return e
		}
		quote, items, e := h.args.Pricing.Quote(ctx, tx, c, true)
		if e != nil {
			return e
		}
		customer, e := h.args.UserRepository.UpsertCustomer(ctx, tx, c.Customer.Name, c.Customer.Phone, c.UserID)
		if e != nil {
			return e
		}
		id, e := h.args.OrderRepository.Insert(ctx, tx, c, customer, quote, hash)
		if e != nil {
			return e
		}
		if e = h.args.OrderRepository.InsertItems(ctx, tx, id, items); e != nil {
			return e
		}
		if e = h.args.InventoryRepository.Reserve(ctx, tx, id, items); e != nil {
			return e
		}
		if e = h.args.PaymentRepository.Insert(ctx, tx, id, c.PaymentMethod, int64(math.Round(quote.Total*100)), c.Key); e != nil {
			return e
		}
		if e = h.args.OrderRepository.History(ctx, tx, id, "", "pending_payment", "", ""); e != nil {
			return e
		}
		out, e = h.args.OrderRepository.GetTx(ctx, tx, id)
		return e
	})
	return out, e
}
