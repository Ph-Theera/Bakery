package checkout

import (
	"bakery/internal/domains/orders"
	catalogout "bakery/internal/ports/port_out/catalog"
	inventoryout "bakery/internal/ports/port_out/inventory"
	storeout "bakery/internal/ports/port_out/store"
	"bakery/internal/shared"
	"context"
	"math"
	"sort"
)

type Pricing struct {
	Products  catalogout.Repository
	Inventory inventoryout.Repository
	Store     storeout.Repository
}

func (s Pricing) Quote(ctx context.Context, tx shared.Transaction, c orders.Command, lock bool) (orders.Quote, []orders.PricedItem, error) {
	settings, e := s.Store.Read(ctx, tx)
	if e != nil {
		return orders.Quote{}, nil, e
	}
	if !settings.Accepting {
		return orders.Quote{}, nil, shared.ErrConflict
	}
	items := append([]orders.Item{}, c.Items...)
	sort.Slice(items, func(i, j int) bool { return items[i].ProductID < items[j].ProductID })
	out := []orders.PricedItem{}
	var subtotal int64
	for _, i := range items {
		p, e := s.Products.Price(ctx, tx, i.ProductID, lock)
		if e != nil {
			return orders.Quote{}, nil, e
		}
		reserved, e := s.Inventory.Reserved(ctx, tx, i.ProductID)
		if e != nil {
			return orders.Quote{}, nil, e
		}
		if int64(i.Quantity) > p.Quantity-reserved {
			return orders.Quote{}, nil, shared.ErrConflict
		}
		if p.Price < 0 || p.Price > 1e11 {
			return orders.Quote{}, nil, shared.ErrInvalid
		}
		subtotal += p.Price * int64(i.Quantity)
		out = append(out, orders.PricedItem{ID: p.ID, Name: p.Name, Quantity: i.Quantity, Price: p.Price})
	}
	fee := int64(math.Round(settings.Fee * 100))
	if c.Fulfillment.Type == "pickup" || settings.FreeThreshold > 0 && subtotal >= int64(math.Round(settings.FreeThreshold*100)) {
		fee = 0
	}
	return orders.Quote{Subtotal: float64(subtotal) / 100, DeliveryFee: float64(fee) / 100, Total: float64(subtotal+fee) / 100}, out, nil
}
